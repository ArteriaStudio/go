package main

import (
	"log"
	"net/http"
	"os"

	"arteria-s.net/api"
)

// HttpdMain Go内蔵のhttpdサーバー機能を用いたプロセスエントリポイント
func HttpdMain(pHandler func(http.ResponseWriter, *http.Request)) {
	// http.Serverインスタンスを生成
	server := &http.Server{
		Addr: ":8443", // Listen on port 8443 for HTTPS.
	}

	// イベントハンドラの登録
	http.HandleFunc("/", pHandler)

	// TLS httpd を起動
	log.Printf("Listening on %s", server.Addr)
	pHome := os.Getenv("APPHOME")
	pCrtFilepath := pHome + "/config/fullchain.pem"
	pKeyFilepath := pHome + "/config/privkey.pem"
	if err := server.ListenAndServeTLS(pCrtFilepath, pKeyFilepath); err != http.ErrServerClosed {
		log.Fatalf("server.ListenAndServeTLS: %v\n", err)
	}
}

// main プロセスエントリポイント
func main() {
	HttpdMain(api.EntryPoint)
}
