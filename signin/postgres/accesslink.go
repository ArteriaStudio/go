package postgres

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/lib/pq"
)

// GetUniqueIdFromSession セッションキーから利用者識別子を獲得
func GetUniqueIdFromSession(pDatabase *sql.DB, pSessionKey string, pUniqueId *string) error {
	var pError error

	if pDatabase != nil {
		pTransaction, pError := Begin(pDatabase)
		if pError == nil {
			defer pTransaction.Commit()

			const pSQL = "SELECT UniqueId FROM TSessions WHERE SessionKey = $1;"
			pRows, pError := pTransaction.Query(pSQL, pSessionKey)
			if pError != nil {
				pTransaction.Rollback()
				log.Printf("FAILED: GetUniqueIdFromSession() SessionKey = %s\n", pSessionKey)
			} else {
				defer pRows.Close()

				if pRows.Next() {
					if pError = pRows.Scan(pUniqueId); pError != nil {
						if pPQError, ok := pError.(*pq.Error); ok {
							pSQLError := pPQError.Code.Name()
							if strings.Compare(pSQLError, "invalid_authorization_specification") == 0 {
								fmt.Println("データベースインスタンスからアクセスを拒否されました。")
								fmt.Println(pError)
							} else {
								fmt.Println("セッション属性情報の入力に失敗しました。[" + pSessionKey + "]")
								fmt.Println(pError)
							}
						}
					}
				}
			}
		}
	}

	return pError
}

// SetClientName 利用者ユニークキーと利用者識別子を記録
func SetClientName(pDatabase *sql.DB, pUniqueId string, eAuthority int, pClientName string) error {
	if pDatabase == nil {
		return fmt.Errorf("FAILED: SetClientName() Authority=[%d], ClientName=[%s]", eAuthority, pClientName)
	}
	pTransaction, pError := pDatabase.Begin()
	if pError != nil {
		return fmt.Errorf("FAILED: SetClientName().Begin Authority=[%d], ClientName=[%s]", eAuthority, pClientName)
	}
	const pSQL = "INSERT INTO TAccessLink (UniqueId, AuthorityId, AccountId) VALUES ($1, $2, $3);"
	pResult, pError := pTransaction.Exec(pSQL, pUniqueId, Authoritys[eAuthority], pClientName)
	if pError != nil {
		log.Printf("INSERT INTO TAccessLink (UniqueId, AuthorityId, AccountId) VALUES ('%s', '%s', '%s');", pUniqueId, Authoritys[eAuthority], pClientName)
		if pPQError, ok := pError.(*pq.Error); ok {
			pSQLError := pPQError.Code.Name()
			if strings.Compare(pSQLError, "unique_violation") == 0 {
				//　既にセッション情報が登録されている
				fmt.Println("既にアクセスリンクが登録されている。[" + pUniqueId + "]")
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
		fmt.Println("アクセスリンクの登録に失敗しました。[" + pUniqueId + "]")
		fmt.Println(pError)
		pTransaction.Rollback()
		return pError
	}

	//
	nRows, _ := pResult.RowsAffected()
	if nRows > 0 {
		pTransaction.Rollback()
		return fmt.Errorf("FAILED: SetClientName().INSERT COUNT=0 Authority=[%d], ClientName=[%s]", eAuthority, pClientName)
	}

	//　セッション情報登録に成功、セッション識別子を返却
	pTransaction.Commit()

	return nil
}

// GetSessionKey 利用者識別子からセッションキーを獲得
func GetSessionKey(pDatabase *sql.DB, eAuthority int, pUniqueId string, pSessionKeys *[]string) error {
	if pDatabase == nil {
		return fmt.Errorf("GetSessionKey(): Invalid Parameter UniqueId=[%s]", pUniqueId)
	}

	pTransaction, pError := Begin(pDatabase)
	if pError != nil {
		return fmt.Errorf("GetSessionKey(): Failed begin transaction UniqueId=[%s]", pUniqueId)
	}
	defer pTransaction.Commit()

	const pSQL = "SELECT SessionKey FROM TSessions WHERE UniqueId = $1;"
	pRows, pError := pTransaction.Query(pSQL, pUniqueId)
	if pError != nil {
		log.Printf("GetSessionKey() AccountId=[%s]\n", pUniqueId)
		pTransaction.Rollback()
		return pError
	}
	defer pRows.Close()

	pKeys := []string{}
	var nRows int = 0
	for pRows.Next() {
		var pSessionKey string
		if pError = pRows.Scan(&pSessionKey); pError != nil {
			if pPQError, ok := pError.(*pq.Error); ok {
				pSQLError := pPQError.Code.Name()
				if strings.Compare(pSQLError, "invalid_authorization_specification") == 0 {
					fmt.Println("データベースインスタンスからアクセスを拒否されました。")
					fmt.Println(pError)
				} else {
					fmt.Println("セッションキー情報の入力に失敗しました。[" + pUniqueId + "] [" + Authoritys[eAuthority] + "]")
					fmt.Println(pError)
				}
			}
			fmt.Print("GetSessionKey(): ")
			fmt.Println(pError)
			return pError
		}
		//log.Printf("pSessionKey[%d]=%s\n", nRows, pSessionKey)
		pKeys = append(pKeys, pSessionKey)

		nRows++
	}
	*pSessionKeys = pKeys

	return nil
}

// GetUniqueId アカウントのユニーク識別子を返却
func GetUniqueId(pDatabase *sql.DB, eAuthority int, pAccountId string, ppUniqueId *string) error {
	if pDatabase == nil {
		return fmt.Errorf("GetUniqueId(): Invalid Parameter AccountId=[%s]", pAccountId)
	}

	pTransaction, pError := Begin(pDatabase)
	if pError != nil {
		return fmt.Errorf("GetUniqueId(): Failed begin transaction AccountId=[%s]", pAccountId)
	}
	defer pTransaction.Commit()

	const pSQL = "SELECT UniqueId FROM TAccessLink WHERE AuthorityId = $1 AND AccountId = $2;"
	pRows, pError := pTransaction.Query(pSQL, Authoritys[eAuthority], pAccountId)
	if pError != nil {
		log.Printf("GetUniqueId() AccountId=[%s], AuthorityId=[%s]\n", pAccountId, Authoritys[eAuthority])
		pTransaction.Rollback()
		return pError
	}
	defer pRows.Close()

	var pUniqueId string = ""
	var nRows int = 0
	for pRows.Next() {
		if pError = pRows.Scan(&pUniqueId); pError != nil {
			if pPQError, ok := pError.(*pq.Error); ok {
				pSQLError := pPQError.Code.Name()
				if strings.Compare(pSQLError, "invalid_authorization_specification") == 0 {
					fmt.Println("データベースインスタンスからアクセスを拒否されました。")
					fmt.Println(pError)
				} else {
					fmt.Println("アカウントユニークキーの入力に失敗しました。[" + pAccountId + "] [" + Authoritys[eAuthority] + "]")
					fmt.Println(pError)
				}
			}
			return pError
		}
		nRows++
	}
	*ppUniqueId = pUniqueId

	return nil
}

// GetAccountId
func GetAccountId(pDatabase *sql.DB, pUniqueId string, eAuthority int, ppAccountId *string) error {
	if pDatabase == nil {
		return fmt.Errorf("GetAccountId(): Invalid Parameter UniqueId=[%s]", pUniqueId)
	}

	pTransaction, pError := Begin(pDatabase)
	if pError != nil {
		return fmt.Errorf("GetAccountId(): Failed begin transaction UniqueId=[%s]", pUniqueId)
	}
	defer pTransaction.Commit()

	const pSQL = "SELECT AccountId FROM TAccessLink WHERE UniqueId = $1 AND AuthorityId = $2;"
	pRows, pError := pTransaction.Query(pSQL, pUniqueId, Authoritys[eAuthority])
	if pError != nil {
		log.Printf("GetAccountId() UniqueId=[%s], AuthorityId=[%s]\n", pUniqueId, Authoritys[eAuthority])
		pTransaction.Rollback()
		return pError
	}
	defer pRows.Close()

	var pAccountId string = ""
	var nRows int = 0
	for pRows.Next() {
		if pError = pRows.Scan(&pAccountId); pError != nil {
			if pPQError, ok := pError.(*pq.Error); ok {
				pSQLError := pPQError.Code.Name()
				if strings.Compare(pSQLError, "invalid_authorization_specification") == 0 {
					fmt.Println("データベースインスタンスからアクセスを拒否されました。")
					fmt.Println(pError)
				} else {
					fmt.Println("アカウントユニークキーの入力に失敗しました。[" + pAccountId + "] [" + Authoritys[eAuthority] + "]")
					fmt.Println(pError)
				}
			}
			return pError
		}
		nRows++
	}
	*ppAccountId = pAccountId

	return nil
}

// UpsertAccessLink
func UpsertAccessLink(pDatabase *sql.DB, pUniqueId string, eAuthority int, pAccountId string) error {
	if pDatabase == nil {
		return fmt.Errorf("FAILED: UpsertAccessLink() Authority=[%d], pAccountId=[%s]", eAuthority, pAccountId)
	}
	pTransaction, pError := pDatabase.Begin()
	if pError != nil {
		return fmt.Errorf("FAILED: UpsertAccessLink().Begin Authority=[%d], pAccountId=[%s]", eAuthority, pAccountId)
	}
	const pSQL = "INSERT INTO TAccessLink (UniqueId, AuthorityId, AccountId) VALUES ($1, $2, $3) ON CONFLICT ON CONSTRAINT taccesslink_pkey DO UPDATE SET CreateStamp = now();"
	pResult, pError := pTransaction.Exec(pSQL, pUniqueId, Authoritys[eAuthority], pAccountId)
	if pError != nil {
		log.Printf("INSERT INTO TAccessLink (UniqueId, AuthorityId, AccountId) VALUES ('%s', '%s', '%s') ON CONFLICT ON CONSTRAINT taccesslink_pkey DO UPDATE SET CreateStamp = now();", pUniqueId, Authoritys[eAuthority], pAccountId)
		if pPQError, ok := pError.(*pq.Error); ok {
			pSQLError := pPQError.Code.Name()
			if strings.Compare(pSQLError, "unique_violation") == 0 {
				//　既にセッション情報が登録されている
				fmt.Println("UpsertAccessLink() 既にアクセスリンクが登録されている。[" + pUniqueId + "]")
				fmt.Println(pError)
				pTransaction.Rollback()
				return pError
			} else if strings.Compare(pSQLError, "invalid_authorization_specification") == 0 {
				fmt.Println("UpsertAccessLink() データベースインスタンスからアクセスを拒否されました。")
				fmt.Println(pError)
				pTransaction.Rollback()
				return pError
			} else {
				fmt.Println("FAILED: SQL= " + pSQL)
				fmt.Println("pSQLError: [" + pSQLError + "]")
			}
		}
		//　セッション情報登録に失敗
		fmt.Println("UpsertAccessLink() アクセスリンクの登録に失敗しました。[" + pUniqueId + "]")
		fmt.Println(pError)
		pTransaction.Rollback()
		return pError
	}

	//
	nRows, _ := pResult.RowsAffected()
	if nRows == 0 {
		pTransaction.Rollback()
		return fmt.Errorf("FAILED: SetClientName().INSERT COUNT=0 Authority=[%d], AccountId=[%s]", eAuthority, pAccountId)
	}

	//　セッション情報登録に成功、セッション識別子を返却
	pTransaction.Commit()

	return nil
}
