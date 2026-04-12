package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"arteria-s.net/postgres"
	"arteria-s.net/signin"
	"golang.org/x/oauth2"
)

// HttpdMain Go内蔵のhttpdサーバー機能を用いたプロセスエントリポイント
func HttpdMain() {
	// http.Serverインスタンスを生成
	server := &http.Server{
		Addr: ":8443", // Listen on port 8443 for HTTPS.
	}

	// イベントハンドラの登録
	http.HandleFunc("/", signin.EntryPoint)

	// TLS httpd を起動
	log.Printf("Listening on %s", server.Addr)
	pHome := os.Getenv("APPHOME")
	pCrtFilepath := pHome + "/config/fullchain.pem"
	pKeyFilepath := pHome + "/config/privkey.pem"
	if err := server.ListenAndServeTLS(pCrtFilepath, pKeyFilepath); err != http.ErrServerClosed {
		log.Fatalf("server.ListenAndServeTLS: %v\n", err)
	}
}

// DebugInsertSession セッションデータ登録（デバッグ用関数）
func DebugInsertSession() {
	//	pServerHostName := "api.arteria-s.net"
	pServerHostName := "localhost"
	pDatabaseName := "abook"
	pDatabase := postgres.Open(pServerHostName, pDatabaseName)
	if pDatabase == nil {
		return
	}
	defer postgres.Close(pDatabase)

	//
	//pApplicationId := "dec981fb-da92-4dde-82a9-8b80bae80071"
	pSessionKey := "123"
	var pToken oauth2.Token
	pToken.AccessToken = "111"
	pToken.RefreshToken = "refresh"
	pToken.Expiry = time.Now()
	pError := postgres.DeleteSessionToken(pDatabase, postgres.AuthorityGoogle, pSessionKey)
	if pError != nil {
		// エラーは無視
	}
	pError = postgres.CreateSessionToken(pDatabase, postgres.AuthorityGoogle, pSessionKey, &pToken)
	if pError != nil {
		fmt.Printf("status = %d\n", pError)
		return
	}
}

func Debug2() {
	// データベースと接続
	pServerHostName := "localhost"
	pDatabaseName := "abook"
	pDatabase := postgres.Open(pServerHostName, pDatabaseName)
	defer postgres.Close(pDatabase)

	//pApplicationId := "dec981fb-da92-4dde-82a9-8b80bae80071"

	//pSessionKey := "123456"

	pUniqueId := ""
	postgres.SetClientName(pDatabase, pUniqueId, postgres.AuthorityGoogle, "client-name")

	pClientName := ""
	pError := postgres.GetAccountId(pDatabase, pUniqueId, postgres.AuthorityGoogle, &pClientName)
	if pError != nil {
		return
	}

	fmt.Println("ClientName: " + pClientName)
}

// main プロセスエントリポイント
func main() {
	HttpdMain()
}
