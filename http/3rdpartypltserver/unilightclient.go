package main

import (
	"fmt"
	"strconv"
	//	"io"
	"io/ioutil"
	"time"
	//"log"
	//"math/rand"
	"crypto/md5"
	"net/http"
	"strings"

	"git.code4.in/mobilegameserver/logging"
	sjson "github.com/bitly/go-simplejson"
)

//var loginUrl = "http://14.17.104.56:8000/zone/clientlog"
var loginUrl = "http://127.0.0.1:7000/httplogin"
var gameid = 170
var zoneid = 301
var gateway1 = 0
var gateway2 = 0
var c = make(chan int)
var shutdown = 0
var cnn = 1
var send = 10
var countMsg = 0

func main() {

	for i := 0; i < cnn; i++ {
		goindex := i
		go connect(goindex)
		logging.Info("go rountion %d", i)
		time.Sleep(10000)
	}
	fmt.Println(<-c)
}

func connect(goindex int) {
	pwd := "qwe123"
	hash := md5.New()
	hash.Write([]byte(pwd))
	sum := fmt.Sprintf("%x", hash.Sum(nil))

	count := fmt.Sprintf("%s: %d", "plattokenlogin", goindex)
	// get serverlist
	serverlist := fmt.Sprintf(`{"do":"request-zone-list", "gameid":%d, "zoneid":301, "data":{"platinfo":{"account":"zwl", "platid":67}}}`, gameid)
	bOk, _ := httpsend(loginUrl, serverlist, count)
	if !bOk {
		logging.Error("httpsend error plat-token-login ")
		return
	}

	//logging.Info("zonelist  %s", string(zonelist))
	// plat-token-login
	plattokenlogin := fmt.Sprintf(`{"do":"plat-token-login", "gameid":%d, "zoneid":301, "data":{"platinfo":{"sign":"%s", "account":"23813", "platid":67}}}`, gameid, sum)
	bOk, token := httpsend(loginUrl, plattokenlogin, count)
	if !bOk {
		logging.Error("httpsend error plat-token-login ")
		return
	}
	logging.Info("plat-token-login %s", string(token))
	js, err := sjson.NewJson(token)
	if err != nil {
		logging.Error("platt-token-login  to json error")
		return
	}
	unigame_plat_key := js.Get("unigame_plat_key").MustString()
	unigame_plat_login := js.Get("unigame_plat_login").MustString()
	uid := js.Get("data").Get("uid").MustString()
	// get userzoneinfo
	data := "{}"
	signurl, dataSend := sendSign(uid, "request-user-zone-info", data, unigame_plat_key, unigame_plat_login, loginUrl, gameid, zoneid)
	bOk, ret := httpsend(signurl, string(dataSend), count)
	if bOk != true {
		logging.Error("httpsend error request-user-zone-info error")
		return
	}

	logging.Info("userzoneinfo  %s", string(ret))
	// select-zone
	data = "{}"
	signurl, dataSend = sendSign(uid, "request-select-zone", data, unigame_plat_key, unigame_plat_login, loginUrl, gameid, zoneid)
	bOk, token = httpsend(signurl, string(dataSend), count)
	if !bOk {
		logging.Error("httpsend error select-zone error")
		return
	}
	js, err = sjson.NewJson(token)
	if err != nil {
		logging.Error("select zone to json error")
		return
	}
	gatewayurl := js.Get("data").Get("gatewayurl").MustString()
	accountid := js.Get("data").Get("zoneuid").MustString()
	uid = fmt.Sprintf("%s", accountid)
	logging.Info("玩家分配的区的uid是%d, %s", accountid, string(token))
	logging.Info("gatewayurl %s", gatewayurl)
	if gatewayurl == "http://14.17.104.56:6502/shen/user/http" {
		gateway2 += 1
	} else {
		gateway1 += 1
	}
	logging.Info("gateway1 %d, gateway2 %d", gateway1, gateway2)
	signurl, dataSend = sendSign(uid, "Cmd.Login_C", "{}", unigame_plat_key, unigame_plat_login, gatewayurl, gameid, zoneid)
	httpsend(signurl, string(dataSend), count)
	// sendTounilight
	signurl, dataSend = sendSign(uid, "Pmd.RequestQueryPlatPointSdkPmd_C", "{}", unigame_plat_key, unigame_plat_login, gatewayurl, gameid, zoneid)
	bOk, token = httpsend(signurl, string(dataSend), count)
	if !bOk {
		logging.Error("httpsend error error")
		return
	}
	logging.Info("积分查询", string(token))

	jsd := sjson.New()
	jsd.Set("goodid", 1)
	jsd.Set("money", 100)
	jsdata, _ := jsd.Encode()
	jsstr := string(jsdata)
	jsstr = strings.Replace(jsstr, "\\", "", -1)
	signurl, dataSend = sendSign(uid, "Pmd.RequestRedeemPlatPointSdkPmd_C", jsstr, unigame_plat_key, unigame_plat_login, gatewayurl, gameid, zoneid)
	bOk, token = httpsend(signurl, string(dataSend), count)
	if !bOk {
		logging.Error("httpsend error error")
		return
	}
	jsd2 := sjson.New()
	jsd2.Set("goodid", 1)
	jsd2.Set("money", 10)
	jsd2.Set("point", 10)
	jsdata, _ = jsd2.Encode()
	jsstr = string(jsdata)
	signurl, dataSend = sendSign(uid, "Pmd.RequestRedeemBackPlatPointSdkPmd_C", jsstr, unigame_plat_key, unigame_plat_login, gatewayurl, gameid, zoneid)
	bOk, token = httpsend(signurl, string(dataSend), count)
	if !bOk {
		logging.Error("httpsend error error")
		return
	}
	logging.Info("带入", string(token))

	c <- 1
}

func sendSign(uid, do, data, unigame_plat_key, unigame_plat_login, url string, gameid, zoneid int) (string, []byte) {
	unigame_plat_timestamp := int(time.Now().Unix())
	jsd := sjson.New()
	jsd.Set("goodid", 1)
	jsd.Set("money", 10)

	js := sjson.New()
	js.Set("do", do)
	js.Set("data", jsd)
	js.Set("unigame_plat_key", unigame_plat_key)
	js.Set("unigame_plat_login", unigame_plat_login)
	js.Set("gameid", gameid)
	js.Set("zoneid", zoneid)
	js.Set("uid", uid)
	js.Set("unigame_plat_timestamp", unigame_plat_timestamp)
	rawdata, _ := js.Encode()

	hash := md5.New()
	timestr := strconv.Itoa(unigame_plat_timestamp)
	hash.Write(append(append(rawdata, ([]byte(timestr))...), unigame_plat_key...))
	sign := fmt.Sprintf("%x", hash.Sum(nil))

	signurl := fmt.Sprintf("%s?unigame_plat_sign=%s", url, sign)
	return signurl, rawdata
}

func httpsend(url, str string, count string) (bool, []byte) {
	resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(str))
	if err == nil {
		ret, _ := ioutil.ReadAll(resp.Body)
		//if err == nil {
		//	fmt.Println("resok", count)
		//}
		defer resp.Body.Close()
		return true, ret
	} else {
		fmt.Println(err, count)
		return false, []byte{}
	}
}
