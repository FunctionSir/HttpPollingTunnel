/*
 * @Author: FunctionSir
 * @License: AGPLv3
 * @Date: 2024-11-03 22:58:10
 * @LastEditTime: 2024-11-10 18:41:19
 * @LastEditors: FunctionSir
 * @Description: -
 * @FilePath: /HttpPollingTunnel/server/main.go
 */

package main

import (
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"gopkg.in/ini.v1"
)

var ClientAuthHash map[string]string
var ClientAuthSalt map[string]string

var BufMap map[string][]byte
var BufMapLock sync.Mutex

var ListenAddr string
var Pattern string

func logFatalln(fatal string) {
	log.Fatalln(fatal)
}

func calcAuthHash(client string, key string) string {
	return fmt.Sprintf("%X", sha512.Sum512([]byte(ClientAuthSalt[client]+key)))
}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	src := r.PostFormValue("src")
	auth := r.PostFormValue("auth")
	if len(src) <= 0 || len(auth) <= 0 {
		w.WriteHeader(http.StatusProxyAuthRequired)
		return
	}
	if calcAuthHash(src, auth) != ClientAuthHash[src] {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	dest := r.PostFormValue("dest")
	data := r.PostFormValue("data")
	if len(data) > 0 {
		raw, err := base64.RawURLEncoding.DecodeString(data)
		if err != nil {
			return
		}
		BufMapLock.Lock()
		BufMap[dest] = append(BufMap[dest], raw...)
		BufMapLock.Unlock()
	}
	BufMapLock.Lock()
	w.Write([]byte(base64.RawURLEncoding.EncodeToString(BufMap[src])))
	BufMap[src] = make([]byte, 0)
	BufMapLock.Unlock()
}

func loadAuthInfo(authFile string) {
	ClientAuthHash = make(map[string]string)
	ClientAuthSalt = make(map[string]string)
	authInfo, err := os.ReadFile(authFile)
	if err != nil {
		logFatalln("Can not load auth file!")
	}
	lines := strings.Split(string(authInfo), "\n")
	for _, l := range lines {
		splited := strings.Split(l, " ")
		if len(splited) != 3 {
			logFatalln("Wrong auth file format!")
		}
		ClientAuthSalt[splited[0]] = splited[1]
		ClientAuthHash[splited[0]] = splited[2]
	}
}

func getConf() {
	if len(os.Args) <= 1 {
		logFatalln("No config file specified!")
	}
	conf, err := ini.Load(os.Args[1])
	if err != nil || !conf.HasSection("server") {
		logFatalln("Can not load config file or wrong config format!")
	}
	sec := conf.Section("server")
	if !sec.HasKey("ListenAddr") || !sec.HasKey("AuthFile") {
		logFatalln("Critical keys not found!")
	}
	ListenAddr = sec.Key("ListenAddr").String()
	loadAuthInfo(sec.Key("AuthFile").String())
	Pattern = "/"
	if sec.HasKey("Pattern") {
		Pattern = sec.Key("Pattern").String()
	}
}
func main() {
	getConf()
	BufMap = make(map[string][]byte)
	http.HandleFunc(Pattern, httpHandler)
	http.ListenAndServe(ListenAddr, nil)
}
