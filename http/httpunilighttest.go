package main

import (
	"crypto/md5"
	"fmt"
	"strconv"
	//	"io"
	"io/ioutil"
	"time"
	//"log"
	//"math/rand"
	"net/http"
	"strings"

	"git.code4.in/mobilegameserver/logging"
	sjson "github.com/bitly/go-simplejson"
)

//var loginUrl = "http://119.63.37.250:7000/httplogin"

//var loginUrl = "http://14.17.104.56:8000/httplogin"
var loginUrl = "http://47.89.42.117:7000/httplogin"
var gameid = 9005
var zoneid = 108
var gateway1 = 0
var gateway2 = 0
var c = make(chan int)
var shutdown = 0
var cnn = 1
var send = 10
var countMsg = 0
var name = "t1_atison1"
var pwd = "123123"

func main() {
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
		time.Sleep(10000000)
	}
	fmt.Println(<-c)
}
func exitfun() {
	c <- 1
}
func connect(goindex int) {
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
	plattokenlogin := fmt.Sprintf(`{"do":"plat-token-login", "gameid":%d, "zoneid":%d, "data":{"platinfo":{"account":"%s", "platid":183, "sign":"%s"}}}`, gameid, zoneid, name, pwd)
	bOk, token := httpsend(loginUrl, plattokenlogin, count)
	if !bOk {
		logging.Error("httpsend error plat-token-login ")
		return
	}
	logging.Info("plat-token-login %s", string(token))
	js, err := sjson.NewJson(token)
	if err != nil || err == nil {
		logging.Error("platt-token-login  to json error")
		exitfun()
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
	// sendTounilight
	/*
		for j := 0; j < send; j += 1 {
			signurl, dataSend := sendSign(uid, "Cmd.UserInfoSynRequestLbyCmd_C", "{}", unigame_plat_key, unigame_plat_login, gatewayurl, gameid, zoneid)
			bOk, token = httpsend(signurl, string(dataSend), count)
			shutdown += 1
			if !bOk {
				logging.Error("httpsend errordquestLbyCmd_c")
				return
			}
			js, err = sjson.NewJson(token)
			if err != nil {
				logging.Error("UserInfoSynRequestLbyCmd_C zone to json error")
				return
			}
			js.Get("data").Get("desc").MustString()
			countMsg += 1
			//logging.Info("rev unilight%s, 第%d个携程中的第%d次访问， 共访问次数%d", desc, goindex, j, countMsg)
		}
			if shutdown > send*cnn-10000 {
				c <- 1
			}
	*/
}

func sendSign(uid, do, data, unigame_plat_key, unigame_plat_login, url string, gameid, zoneid int) (string, []byte) {
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
