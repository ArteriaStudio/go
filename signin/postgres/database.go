package postgres

import (
	"database/sql"
	"fmt"
	"log"

	// Postgres Database Driver
	_ "github.com/lib/pq"
)

// handleError データベースエラーハンドラ
func handleError(pErr error) {
	log.Fatal(pErr)
}

// Open データベース接続プールを生成
func Open(pServerHostName string, pDatabaseName string) (pDatabase *sql.DB) {
	pParams := "host=" + pServerHostName + " dbname=" + pDatabaseName + " sslmode=disable user=aploper"
	pDatabase, pErr := sql.Open("postgres", pParams)
	if pErr != nil {
		fmt.Printf("FAILED：データベースインスタンス（%s:%s）に接続できません。\n", pServerHostName, pDatabaseName)
		handleError(pErr)
	} else {
		pDatabase.Stats()
	}

	return
}

// Close データベース接続プールを削除
func Close(pDatabase *sql.DB) {
	pError := pDatabase.Close()
	if pError != nil {
		fmt.Println("FAILED：データベースインスタンスから切断できませんでした。")
	}
}
