package database

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
	pParams := "host=" + pServerHostName + " dbname=" + pDatabaseName + " sslmode=disable user=api"
	pDatabase, pErr := sql.Open("postgres", pParams)
	if pErr != nil {
		fmt.Println(pErr)
		return nil
	}
	if pDatabase != nil {
		pError := pDatabase.Ping()
		if pError != nil {
			fmt.Printf("FAILED: インスタンス（%s:%s）に接続できませんでした。\n", pServerHostName, pDatabaseName)
			return nil
		}
		fmt.Printf("SUCCESS：データベースインスタンス（%s:%s）に接続しました。\n", pServerHostName, pDatabaseName)
		pDatabase.Stats()
	} else {
		handleError(pErr)
	}

	return pDatabase
}

// Close データベース接続プールを削除
func Close(pDatabase *sql.DB) {
	pDatabase.Close()
	fmt.Println("SUCCESS：データベースインスタンスから切断しました。")
}
