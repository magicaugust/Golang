package main

import (
	"crypto/md5"
	"fmt"
	"strconv"
	"io/ioutil"
	"time"
	"net/http"
	"strings"
	sjson "github.com/bitly/go-simplejson"
	"git.code4.in/mobilegameserver/logging"
)

var loginUrl = "http://server.login.bwgame.com.cn:8000/httplogin"
var gameid = 4000 
var zoneid = 301 
var gateway1 = 0
var gateway2 = 0
var c = make(chan int) 
var shutdown = 0
var cnn = 1 
var send =  1
var countMsg = 0


func main(){
	/*
	filePath, _ := filepath.Abs(os.Args[0])
	if os.Getppid() != 1 {
			logging.Info("server start as daemon:%s,%v", filePath, os.Args[1:])
			cmd := exec.Command(filePath, os.Args[1:]...)
			cmd.Start()
			os.Exit(0)
		}
	*/


	for i := 0; i < cnn; i++ {
		goindex := i 
		go connect(goindex)
		logging.Info("go rountion %d", i)
		time.Sleep(10000)
	}
	fmt.Println(<-c)
}

func connect(goindex int) {
	count := fmt.Sprintf("%s: %d", "plattokenlogin", goindex)

	plattokenlogin := fmt.Sprintf(`{"do":"plat-token-login", "gameid":%d, "zoneid":301, "data":{"platinfo":{"account":"zwl", "platid":0}}}`, gameid)
	bOk, token := httpsend(loginUrl, plattokenlogin, count)
	if !bOk {
		logging.Error("httpsend error plat-token-login ")
		return 
	}
	js, err := sjson.NewJson(token)
	if err != nil {
		logging.Error("platt-token-login  to json error")
		return 
	}

	unigame_plat_key := js.Get("unigame_plat_key").MustString()
	unigame_plat_login := js.Get("unigame_plat_login").MustString()
	uid := js.Get("data").Get("uid").MustString()
	// select zone 

	data := "{}"
	signurl, dataSend := sendSign(uid, "request-select-zone", data, unigame_plat_key, unigame_plat_login, loginUrl, gameid, zoneid)
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
	logging.Info("gatewayurl %s revdata %s", gatewayurl, string(token))

	signurl, dataSend = sendSign(uid, "Pmd.WebSocketForwardUserPmd_C", "{}", unigame_plat_key, unigame_plat_login, gatewayurl, gameid, zoneid)
	bOk, token = httpsend(signurl, string(dataSend), count)
	shutdown += 1
	if !bOk {
		logging.Error("httpsend errordquestLbyCmd_c")
		return 
	}
	logging.Info("websocketForwardUserPmd_C %s", string(token))
	js, err = sjson.NewJson(token)
	if err != nil {
		logging.Error("UserInfoSynRequestLbyCmd_C zone to json error")
		return 
	}
	js.Get("data").Get("desc").MustString()
	countMsg += 1
	//logging.Info("rev unilight%s, 第%d个携程中的第%d次访问， 共访问次数%d", desc, goindex, j, countMsg)
	if shutdown > send*cnn-10000 {
		c <- 1	
	}
}

func sendSign(uid, do, data, unigame_plat_key, unigame_plat_login, url string, gameid, zoneid int)(string, []byte){
	unigame_plat_timestamp := int(time.Now().Unix())
	js := sjson.New()
	js.Set("do", do)
	js.Set("data", data)
	js.Set("unigame_plat_key", unigame_plat_key)
	js.Set("unigame_plat_login", unigame_plat_login)
	js.Set("gameid", gameid)
	js.Set("zoneid", zoneid)
	js.Set("uid", uid)
	js.Set("unigame_plat_timestamp", unigame_plat_timestamp)
	rawdata,_ := js.Encode()

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
