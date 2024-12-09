// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ws

import (
	"log"
	"net/http"
	"sync"
)

//
//var addr = flag.String("addr", ":8080", "http service address")
//
//func serveHome(w http.ResponseWriter, r *http.Request) {
//	log.Println(r.URL)
//	if r.URL.Path != "/" {
//		http.Error(w, "Not found", http.StatusNotFound)
//		return
//	}
//	if r.Method != http.MethodGet {
//		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
//		return
//	}
//	http.ServeFile(w, r, "home.html")
//}
//
//func main() {
//	flag.Parse()
//	hub := newHub()
//	go hub.run()
//	http.HandleFunc("/", serveHome)
//	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
//		serveWs(hub, w, r)
//	})
//	err := http.ListenAndServe(*addr, nil)
//	if err != nil {
//		log.Fatal("ListenAndServe: ", err)
//	}
//}

var wsOnce sync.Once
var THub *TradeHub

func InitTradeHub(symbols map[string]int64) {
	THub = newHub(symbols)
}
func WsHandle(w http.ResponseWriter, r *http.Request) {
	wsOnce.Do(func() {
		go THub.run()
		log.Println("trade hub start*******")
	})
	serveWs(THub, w, r)
}
