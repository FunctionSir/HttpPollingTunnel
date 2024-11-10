/*
 * @Author: FunctionSir
 * @License: AGPLv3
 * @Date: 2024-11-02 23:58:18
 * @LastEditTime: 2024-11-10 20:26:05
 * @LastEditors: FunctionSir
 * @Description: -
 * @FilePath: /HttpPollingTunnel/client/main.go
 */

package main

import (
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/songgao/water"
	"gopkg.in/ini.v1"
)

var Buf []byte
var Backlog int
var BufLock sync.Mutex

var ErrCnt int
var ErrCntLock sync.Mutex

var TunName string

var Server string
var Src string
var Dest string
var Auth string

var AutoRelax bool
var LastHaveData int64
var LastHaveDataLock sync.Mutex
var ExchangeGap int

var RefreshGap int

func logInfoln(info string) {
	log.Println(info)
}

func logInfof(info string, v ...any) {
	log.Printf(info, v...)
}

func logWarnln(warn string) {
	log.Println(warn)
}

func logFatalln(fatal string) {
	log.Fatalln(fatal)
}

func IncErrCnt() {
	ErrCntLock.Lock()
	ErrCnt++
	ErrCntLock.Unlock()
}

func GetErrCnt() int {
	var curErrCnt int
	ErrCntLock.Lock()
	curErrCnt = ErrCnt
	ErrCntLock.Unlock()
	return curErrCnt
}

func valueToBool(s string) bool {
	if s == "true" || s == "True" || s == "T" || s == "t" || s == "1" {
		return true
	}
	return false
}

func doPost(srv, src, auth, dest, data string) (string, error) {
	toPost := url.Values{"src": {src}, "dest": {dest}, "auth": {auth}, "data": {data}}
	resp, err := http.PostForm(srv, toPost)
	if err != nil {
		IncErrCnt()
		return "", err
	}
	defer resp.Body.Close()
	tmpBuf, err := io.ReadAll(resp.Body)
	if err != nil {
		IncErrCnt()
		return "", err
	}
	return string(tmpBuf), nil
}

func readFromTun(tun *water.Interface) {
	tmpBuf := make([]byte, 1500*4)
	for {
		n, err := tun.Read(tmpBuf)
		if err != nil {
			IncErrCnt()
			logInfoln("Error occurred when trying to read from a tun: " + err.Error())
			continue
		}
		BufLock.Lock()
		Buf = append(Buf, tmpBuf[:n]...)
		Backlog++
		BufLock.Unlock()
	}
}

func updateLastHaveData() {
	if AutoRelax {
		LastHaveDataLock.Lock()
		LastHaveData = time.Now().UnixMilli()
		LastHaveDataLock.Unlock()
	}
}

func getLastHaveData() int64 {
	LastHaveDataLock.Lock()
	cur := LastHaveData
	LastHaveDataLock.Unlock()
	return cur
}

func infoExchange(tun *water.Interface) {
	for {
		time.Sleep(time.Duration(ExchangeGap) * time.Microsecond)
		if AutoRelax {
			curLastHaveData := getLastHaveData()
			now := time.Now().UnixMilli()
			for AutoRelax && len(Buf) == 0 && now-curLastHaveData >= 1000 && now-curLastHaveData <= 5000 {
				//! Relax !//
			}
		}
		BufLock.Lock()
		if AutoRelax {
			if len(Buf) != 0 {
				go updateLastHaveData()
			}
		}
		encoded := base64.RawURLEncoding.EncodeToString(Buf)
		Buf = make([]byte, 0)
		Backlog = 0
		BufLock.Unlock()
		raw, err := doPost(Server, Src, Auth, Dest, encoded)
		if err != nil {
			IncErrCnt()
			logWarnln("Error occurred when posting: " + err.Error())
			continue
		}
		if len(raw) > 0 {
			if AutoRelax {
				go updateLastHaveData()
			}
			decoded, err := base64.RawURLEncoding.DecodeString(raw)
			if err != nil {
				IncErrCnt()
				logWarnln("Error occurred when decode: " + err.Error())
				continue
			}
			tun.Write(decoded)
		}
	}
}

func getConf() {
	if len(os.Args) <= 1 {
		logFatalln("No config file specified!")
	}
	conf, err := ini.Load(os.Args[1])
	if err != nil || !conf.HasSection("client") {
		logFatalln("Can not load config file or wrong config format!")
	}
	sec := conf.Section("client")
	if !sec.HasKey("Server") || !sec.HasKey("Src") || !sec.HasKey("Dest") || !sec.HasKey("Auth") {
		logFatalln("Critical keys not found!")
	}
	Server = sec.Key("Server").String()
	Src = sec.Key("Src").String()
	Dest = sec.Key("Dest").String()
	Auth = sec.Key("Auth").String()
	if sec.HasKey("TunName") {
		TunName = sec.Key("TunName").String()
	} else {
		TunName = "ht0"
	}
	AutoRelax = false
	if sec.HasKey("AutoRelax") {
		AutoRelax = valueToBool(sec.Key("AutoRelax").String())
	}
	ExchangeGap = 10
	if sec.HasKey("ExchangeGap") {
		tmpExchangeGap, err := strconv.Atoi(sec.Key("ExchangeGap").String())
		if err != nil {
			logFatalln("Wrong format in a value.")
		}
		ExchangeGap = tmpExchangeGap
	}
	RefreshGap = 15
	if sec.HasKey("RefreshGap") {
		tmpRefreshGap, err := strconv.Atoi(sec.Key("RefreshGap").String())
		if err != nil {
			logFatalln("Wrong format in a value.")
		}
		RefreshGap = tmpRefreshGap
	}
}

func main() {
	getConf()
	tunCfg := water.Config{DeviceType: water.TUN, PlatformSpecificParams: water.PlatformSpecificParams{Name: TunName}}
	tun, err := water.New(tunCfg)
	if err != nil {
		panic(err)
	}
	Buf = make([]byte, 0)
	go readFromTun(tun)
	updateLastHaveData()
	go infoExchange(tun)
	for {
		logInfof("Buffered: %d, Backlog: %d, Error: %d\n", len(Buf), Backlog, GetErrCnt())
		time.Sleep(time.Duration(RefreshGap) * time.Second)
	}
}
