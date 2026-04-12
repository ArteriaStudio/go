package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

// AddLoginHistory ログイン成功履歴を記録
func AddLoginHistory(pDatabase *sql.DB, pUniqueId string, eAuthority int, pAccountId string, pClientIP string) error {
	if pDatabase == nil {
		return fmt.Errorf("AddLoginHistory(): Database is null")
	}
	pTransaction, pError := pDatabase.Begin()
	if pError != nil {
		return pError
	}

	const pSQL = "INSERT INTO HLogins (UniqueId, AuthorityId, AccountId, ClientIP) VALUES ($1, $2, $3, $4);"
	//log.Printf("INSERT INTO HLogins (UniqueId, AuthorityId, AccountId, ClientIP) VALUES ('%s', '%s', '%s', '%s');", pUniqueId, Authoritys[eAuthority], pAccountId, pClientIP)
	pResult, pError := pTransaction.Exec(pSQL, pUniqueId, Authoritys[eAuthority], pAccountId, pClientIP)
	if pError != nil {
		if pPQError, ok := pError.(*pq.Error); ok {
			pSQLError := pPQError.Code.Name()
			if strings.Compare(pSQLError, "unique_violation") == 0 {
				//　既にセッション情報が登録されている
				fmt.Println("既にセッション情報が登録されている。[" + pAccountId + "]")
				fmt.Println(pError)
				pTransaction.Rollback()
				return pError
			} else if strings.Compare(pSQLError, "invalid_authorization_specification") == 0 {
				fmt.Println("データベースインスタンスからアクセスを拒否されました。")
				fmt.Println(pError)
				pTransaction.Rollback()
				return pError
			} else {
				fmt.Println("FAILED: SQL= " + pSQL)
				fmt.Println("pSQLError: [" + pSQLError + "]")
			}
		}
		//　セッション情報登録に失敗
		fmt.Println("セッション情報登録に失敗しました。[" + pAccountId + "]")
		fmt.Println(pError)
		pTransaction.Rollback()
		return pError
	}

	//
	nRows, pError := pResult.RowsAffected()
	if nRows == 0 {
		pTransaction.Rollback()
		return pError
	}

	//　セッション情報登録に成功、セッション識別子を返却
	pTransaction.Commit()

	return nil
}

// IsExistsHistory 過去に通信したことがあるグローバルIPアドレスであるかを検査
func IsExistsHistory(pDatabase *sql.DB, eAuthority int, pAccountId string, pClientIP string) (int, error) {
	if pDatabase == nil {
		return 0, fmt.Errorf("AddIsExistsHistoryLoginHistory(): Database is null")
	}
	pTransaction, pError := Begin(pDatabase)
	if pError != nil {
		return 0, pError
	}
	defer pTransaction.Commit()

	const pSQL = "SELECT COUNT(*) FROM HLogins WHERE AccountId = $1 AND ClientIP = $2 AND AuthorityId = $3;"
	//log.Printf("SELECT COUNT(*) FROM HLogins WHERE AccountId = '%s' AND ClientIP = '%s' AND AuthorityId = '%s';", pAccountId, pClientIP, Authoritys[eAuthority])
	pRows, pError := pTransaction.Query(pSQL, pAccountId, pClientIP, Authoritys[eAuthority])
	if pError != nil {
		pTransaction.Rollback()
		return 0, pError
	}
	defer pRows.Close()

	var nRows int = 0
	if pRows.Next() {
		if pError = pRows.Scan(&nRows); pError != nil {
			if pPQError, ok := pError.(*pq.Error); ok {
				pSQLError := pPQError.Code.Name()
				if strings.Compare(pSQLError, "invalid_authorization_specification") == 0 {
					fmt.Println("データベースインスタンスからアクセスを拒否されました。")
					fmt.Println(pError)
				} else {
					fmt.Println("ログイン成功履歴の入力に失敗しました。[" + pAccountId + "] [" + pClientIP + "] [" + Authoritys[eAuthority] + "]")
					fmt.Println(pError)
				}
			}
			return 0, pError
		}
	}

	return nRows, nil
}
