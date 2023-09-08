package database

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

// Transaction トランザクション
type Transaction struct {
	pTx *sql.Tx
}

// Begin トランザクションを開始
func Begin(pDatabase *sql.DB) (pTransaction Transaction, pError error) {
	pTx, pError := pDatabase.Begin()
	if pError != nil {
		//　トランザクションの開始に失敗
		return
	}
	pTransaction.pTx = pTx

	return
}

// Rollback ロールバック
func (pTransaction *Transaction) Rollback() bool {
	pErr := pTransaction.pTx.Rollback()
	if pErr != nil {
		return (false)
	}

	return (true)
}

// Commit ロールバック
func (pTransaction *Transaction) Commit() bool {
	pErr := pTransaction.pTx.Commit()
	if pErr != nil {
		return (false)
	}

	return (true)
}

// Execute SQLを実行（DML）
func (pTransaction *Transaction) Execute(pSQL string, args ...interface{}) bool {
	pResult, pErr := pTransaction.pTx.Exec(pSQL, args...)
	if pErr != nil {
		if pPQError, ok := pErr.(*pq.Error); ok {
			pSQLError := pPQError.Code.Name()
			if strings.Compare(pSQLError, "unique_violation") == 0 {
				//　ユニークキー違反
				fmt.Println(pErr)
				pTransaction.pTx.Rollback()
				return (false)
			} else if strings.Compare(pSQLError, "invalid_authorization_specification") == 0 {
				fmt.Println(pErr)
				pTransaction.pTx.Rollback()
				return (false)
			} else {
				fmt.Println("pSQLError: [" + pSQLError + "]")
			}
		}
		//　セッション情報登録に失敗
		fmt.Println(pErr)
		pTransaction.pTx.Rollback()
		return (false)
	}
	fmt.Println(pResult)

	return (true)
}

// Query カーソルを開く
func (pTransaction *Transaction) Query(pSQL string, args ...interface{}) (pRows *sql.Rows, pErr error) {
	pRows, pErr = pTransaction.pTx.Query(pSQL, args...)
	if pErr != nil {
		fmt.Println(pErr)
		return
	}

	return
}
