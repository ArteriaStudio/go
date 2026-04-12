package postgres

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/lib/pq"
)

// TSubscribeId
type TSubscribeId struct {
	ResourceId  string // リソース識別子
	SubscribeId string // サブスクリプション識別子
}

// CreateSubscribe 変更通知情報を生成
func CreateSubscribe(pDatabase *sql.DB, pUniqueId string, eAuthority int, pAccountId string, pResourceId string, pClientState string, pExpireTimeStamp time.Time, pSubscribeId string) error {
	// トランザクションを開始
	pTx, pError := pDatabase.Begin()
	if pError != nil {
		// トランザクションの開始に失敗
		return fmt.Errorf("CreateSubscribe(): Failure transaction begin. SubscribeId=[%s], Error=[%s]", pSubscribeId, pError.Error())
	}
	pExpireDateTime := pExpireTimeStamp.Format(time.RFC1123Z)

	// サブスクリプションの記録
	const pSQL = "INSERT INTO TWebHooks (UniqueId, AuthorityId, AccountId, ResourceId, SubscribeId, ClientState, ExpireStamp) VALUES ($1, $2, $3, $4, $5, $6, $7);"
	pResult, pError := pTx.Exec(pSQL, pUniqueId, Authoritys[eAuthority], pAccountId, pResourceId, pSubscribeId, pClientState, pExpireDateTime)
	if pError != nil {
		log.Printf("AccountId: %s, Authoritys: %s(%d), ResourceId: %s, SubscribeId: %s, ClientState: %s, ExpireDateTime: %s\n", pAccountId, Authoritys[eAuthority], eAuthority, pResourceId, pSubscribeId, pClientState, pExpireDateTime)
		pTx.Rollback()
		if pPQError, ok := pError.(*pq.Error); ok {
			pSQLError := pPQError.Code.Name()
			if strings.Compare(pSQLError, "unique_violation") == 0 {
				//　既にサブスクリプションの記録されている
				return fmt.Errorf("CreateSubscribe(): 既にサブスクリプションの記録されている。（登録）SubscribeID=[%s], Error=[%s]", pSubscribeId, pError.Error())
			} else if strings.Compare(pSQLError, "invalid_authorization_specification") == 0 {
				return fmt.Errorf("CreateSubscribe(): データベースインスタンスからアクセスを拒否されました。SubscribeID=[%s], Error=[%s]", pSubscribeId, pError.Error())
			}
		}
		// サブスクリプションの記録に失敗
		return fmt.Errorf("CreateSubscribe(): サブスクリプションの記録に失敗しました。SubscribeID=[%s], Error=[%s]", pSubscribeId, pError.Error())
	}
	// サブスクリプションの記録に成功、変更通知識別子を返却
	nRows, _ := pResult.RowsAffected()
	if nRows == 0 {
		return fmt.Errorf("CreateSubscribe(): Failed INSERT. Count=(%d)", nRows)
	}

	pTx.Commit()

	return nil
}

// UpdateSubscribe 変更通知情報を更新
func UpdateSubscribe(pDatabase *sql.DB, eAuthority int, pAccountId string, pResourceId string, pClientState string, pSubscribeId string, pExpireStamp string) (int64, error) {
	// トランザクションを開始
	pTx, pError := pDatabase.Begin()
	if pError != nil {
		// トランザクションの開始に失敗
		return 0, fmt.Errorf("UpdateSubscribe(): Failure transaction begin. SubscribeId=[%s], Error=[%s]", pSubscribeId, pError.Error())
	}

	// サブスクリプションの記録
	const pSQL = "UPDATE TWebHooks SET SubscribeId = $1, ExpireStamp = $2 WHERE AuthorityId = $4 AND AccountId = $5 AND ResourceId = $6 AND ClientState = $7;"
	pResult, pError := pTx.Exec(pSQL, pSubscribeId, pExpireStamp, Authoritys[eAuthority], pAccountId, pResourceId, pClientState)
	if pError != nil {
		if pPQError, ok := pError.(*pq.Error); ok {
			pSQLError := pPQError.Code.Name()
			if strings.Compare(pSQLError, "unique_violation") == 0 {
				//　既にサブスクリプションの記録されている
				pTx.Rollback()
				return 0, fmt.Errorf("UpdateSubscribe(): 既にサブスクリプションの記録されている。（更新）SubscribeID=[%s], Error=[%s]", pSubscribeId, pError.Error())
			} else if strings.Compare(pSQLError, "invalid_authorization_specification") == 0 {
				pTx.Rollback()
				return 0, fmt.Errorf("UpdateSubscribe(): データベースインスタンスからアクセスを拒否されました。SubscribeID=[%s], Error=[%s]", pSubscribeId, pError.Error())
			}
			return 0, fmt.Errorf("UpdateSubscribe(): 行の更新に失敗しました。SubscribeID=[%s], Error=[%s]", pSubscribeId, pError.Error())
		}
		// サブスクリプションの記録に失敗
		pTx.Rollback()
		return 0, fmt.Errorf("UpdateSubscribe(): サブスクリプションの記録に失敗しました。SubscribeID=[%s], Error=[%s]", pSubscribeId, pError.Error())
	}
	nRows, _ := pResult.RowsAffected()

	// サブスクリプションの記録に成功、変更通知識別子を返却
	fmt.Printf("SUCCESS: UPDATE TWebHooks: %d\n", nRows)

	pTx.Commit()

	return nRows, nil
}

// DeleteSubscribe 変更通知情報を生成
func DeleteSubscribe(pDatabase *sql.DB, eAuthority int, pAccountId string, pSubscribeId string) error {
	if pDatabase == nil {
		return fmt.Errorf("DeleteSubscribe(): Invalid parameter at Database. SubscribeId=[%s]", pSubscribeId)
	}
	pTransaction, pError := Begin(pDatabase)
	if pError != nil {
		return fmt.Errorf("DeleteSubscribe(): Failure transaction begin. SubscribeId=[%s]", pSubscribeId)
	}
	const pSQL = "DELETE FROM TWebHooks WHERE AuthorityId = $1 AND AccountId = $2 AND SubscribeId = $3;"
	if !pTransaction.Execute(pSQL, Authoritys[eAuthority], pAccountId, pSubscribeId) {
		pTransaction.Rollback()
		return fmt.Errorf("DeleteSubscribe(): Failure SELECT Query.SQL=[%s]", pSQL)
	}
	log.Println("DELETE FROM TWebHooks; ID=" + pSubscribeId)

	pTransaction.Commit()

	return nil
}

// LookupSubscribe 変更通知情報から変更通知識別子に対応するレコードを検索
func LookupSubscribe(pDatabase *sql.DB, eAuthority int, pSubscribeId string, pClientState string, pUniqueId *string, pAccountId *string, pExpireSubscribe *time.Time) (int, error) {
	if pDatabase == nil {
		return 0, fmt.Errorf("LookupSubscribe(): Invalid parameter at Database. SubscribeId=[%s]", pSubscribeId)
	}
	pTransaction, pError := Begin(pDatabase)
	if pError != nil {
		return 0, fmt.Errorf("LookupSubscribe(): Failure transaction begin. SubscribeId=[%s]", pSubscribeId)
	}
	defer pTransaction.Commit()

	const pSQL = "SELECT UniqueId, AccountId, ResourceId, ExpireStamp FROM TWebHooks WHERE AuthorityId = $1 AND SubscribeId = $2 AND ClientState = $3;"
	pRows, pError := pTransaction.Query(pSQL, Authoritys[eAuthority], pSubscribeId, pClientState)
	if pError != nil {
		pTransaction.Rollback()
		return 0, fmt.Errorf("LookupSubscribe(): Failure SELECT Query.SQL=[%s]", pSQL)
	}
	defer pRows.Close()

	var nRows int = 0
	for pRows.Next() {
		var pResourceId string
		var pExpireStamp string
		if pError = pRows.Scan(pUniqueId, pAccountId, &pResourceId, &pExpireStamp); pError != nil {
			if pPQError, ok := pError.(*pq.Error); ok {
				pSQLError := pPQError.Code.Name()
				if strings.Compare(pSQLError, "invalid_authorization_specification") == 0 {
					return 0, fmt.Errorf("LookupSubscribe(): データベースインスタンスからアクセスを拒否されました。SQLError=[%s]", pSQLError)
				}
				return 0, fmt.Errorf("LookupSubscribe(): サブスクリプションの入力に失敗しました。pSubscribeId=[%s],ExpireStamp=[%s],SQLError=[%s]", pSubscribeId, pExpireSubscribe.Format(time.RFC3339), pSQLError)
			}
			return 0, fmt.Errorf("LookupSubscribe(): サブスクリプションの入力に失敗しました。pSubscribeId=[%s],ExpireStamp=[%s]", pSubscribeId, pExpireSubscribe.Format(time.RFC3339))
		} else {
			nRows++
			pStamp, pError := time.Parse(time.RFC3339, pExpireStamp)
			if pError != nil {
				return 0, fmt.Errorf("Parse(): 日時のパースに失敗しました。pSubscribeId=[%s],ExpireStamp=[%s]", pSubscribeId, pExpireStamp)
			} else {
				*pExpireSubscribe = pStamp
			}
			break
		}
	}

	return nRows, nil
}

// ListupSubscribe サブスクリプション記録情報から指定条件に一致するレコードを返却
func ListupSubscribe(pDatabase *sql.DB, eAuthority int, pAccountId string) ([]TSubscribeId, error) {
	if pDatabase == nil {
		return nil, fmt.Errorf("ListupSubscribe(): Invalid parameter at Database. AccountId=[%s]", pAccountId)
	}
	pTransaction, pError := Begin(pDatabase)
	if pError != nil {
		return nil, fmt.Errorf("ListupSubscribe(): Failure transaction begin. AccountId=[%s]", pAccountId)
	}
	defer pTransaction.Commit()

	const pSQL = "SELECT ResourceId, SubscribeId FROM TWebHooks WHERE AuthorityId = $1 AND AccountId = $2;"
	pRows, pError := pTransaction.Query(pSQL, Authoritys[eAuthority], pAccountId)
	if pError != nil {
		pTransaction.Rollback()
		return nil, fmt.Errorf("ListupSubscribe(): Failure SELECT Query.SQL=[%s]", pSQL)
	}
	defer pRows.Close()

	var pResults []TSubscribeId
	for pRows.Next() {
		var pResult TSubscribeId
		if pError = pRows.Scan(&pResult.ResourceId, &pResult.SubscribeId); pError != nil {
			if pPQError, ok := pError.(*pq.Error); ok {
				pSQLError := pPQError.Code.Name()
				if strings.Compare(pSQLError, "invalid_authorization_specification") == 0 {
					return nil, fmt.Errorf("ListupSubscribe(): データベースインスタンスからアクセスを拒否されました。SQLError=[%s]", pSQLError)
				}
				return nil, fmt.Errorf("ListupSubscribe(): サブスクリプションの入力に失敗しました。pAccountId=[%s],SQLError=[%s]", pAccountId, pSQLError)
			}
			return nil, fmt.Errorf("ListupSubscribe(): サブスクリプションの入力に失敗しました。pAccountId=[%s]", pAccountId)
		} else {
			pResults = append(pResults, pResult)
		}
	}

	return pResults, nil
}

// UpdateSubscribe 変更通知情報を更新
func UpdateSyncTimestamp(pDatabase *sql.DB, eAuthority int, pSubscribeId string, pSyncTimestamp string) (int64, error) {
	// トランザクションを開始
	pTx, pError := pDatabase.Begin()
	if pError != nil {
		// トランザクションの開始に失敗
		return 0, fmt.Errorf("UpdateSyncTimestamp(): Failure transaction begin. SubscribeId=[%s], Error=[%s]", pSubscribeId, pError.Error())
	}

	// サブスクリプションの記録
	const pSQL = "UPDATE TWebHooks SET SyncTimestamp = $1 WHERE AuthorityId = $2 AND AccountId = $3 AND SubscribeId = $4;"
	pResult, pError := pTx.Exec(pSQL, pSyncTimestamp, Authoritys[eAuthority], pSubscribeId)
	if pError != nil {
		if pPQError, ok := pError.(*pq.Error); ok {
			pSQLError := pPQError.Code.Name()
			if strings.Compare(pSQLError, "unique_violation") == 0 {
				//　既にサブスクリプションの記録されている
				pTx.Rollback()
				return 0, fmt.Errorf("UpdateSyncTimestamp(): 既にサブスクリプションの記録されている。（更新）SubscribeID=[%s], Error=[%s]", pSubscribeId, pError.Error())
			} else if strings.Compare(pSQLError, "invalid_authorization_specification") == 0 {
				pTx.Rollback()
				return 0, fmt.Errorf("UpdateSyncTimestamp(): データベースインスタンスからアクセスを拒否されました。SubscribeID=[%s], Error=[%s]", pSubscribeId, pError.Error())
			}
			return 0, fmt.Errorf("UpdateSyncTimestamp(): 行の更新に失敗しました。SubscribeID=[%s], Error=[%s]", pSubscribeId, pError.Error())
		}
		// サブスクリプションの記録に失敗
		pTx.Rollback()
		return 0, fmt.Errorf("UpdateSyncTimestamp(): サブスクリプションの記録に失敗しました。SubscribeID=[%s], Error=[%s]", pSubscribeId, pError.Error())
	}
	nRows, _ := pResult.RowsAffected()

	// サブスクリプションの記録に成功、変更通知識別子を返却
	fmt.Printf("SUCCESS: UPDATE TWebHooks SET SyncTimestamp: %d\n", nRows)

	pTx.Commit()

	return nRows, nil
}

// UpdateSubscribeTimestamp サブスクリプションの有効期限を更新
func UpdateSubscribeTimestamp(pDatabase *sql.DB, eAuthority int, pSubscribeId string, pExpireTimestamp time.Time) (int64, error) {
	// トランザクションを開始
	pTx, pError := pDatabase.Begin()
	if pError != nil {
		// トランザクションの開始に失敗
		return 0, fmt.Errorf("UpdateSyncTimestamp(): Failure transaction begin. SubscribeId=[%s], Error=[%s]", pSubscribeId, pError.Error())
	}

	// サブスクリプションの記録
	pExpireStamp := pExpireTimestamp.Format(time.RFC3339)
	const pSQL = "UPDATE TWebHooks SET ExpireStamp = $1 WHERE AuthorityId = $2 AND SubscribeId = $3;"
	pResult, pError := pTx.Exec(pSQL, pExpireStamp, Authoritys[eAuthority], pSubscribeId)
	if pError != nil {
		if pPQError, ok := pError.(*pq.Error); ok {
			pSQLError := pPQError.Code.Name()
			if strings.Compare(pSQLError, "unique_violation") == 0 {
				//　既にサブスクリプションの記録されている
				pTx.Rollback()
				return 0, fmt.Errorf("UpdateSyncTimestamp(): 既にサブスクリプションの記録されている。（更新）SubscribeID=[%s], Error=[%s]", pSubscribeId, pError.Error())
			} else if strings.Compare(pSQLError, "invalid_authorization_specification") == 0 {
				pTx.Rollback()
				return 0, fmt.Errorf("UpdateSyncTimestamp(): データベースインスタンスからアクセスを拒否されました。SubscribeID=[%s], Error=[%s]", pSubscribeId, pError.Error())
			}
			return 0, fmt.Errorf("UpdateSyncTimestamp(): 行の更新に失敗しました。SubscribeID=[%s], Error=[%s]", pSubscribeId, pError.Error())
		}
		// サブスクリプションの記録に失敗
		pTx.Rollback()
		return 0, fmt.Errorf("UpdateSyncTimestamp(): サブスクリプションの記録に失敗しました。SubscribeID=[%s], Error=[%s]", pSubscribeId, pError.Error())
	}
	nRows, _ := pResult.RowsAffected()

	// サブスクリプションの記録に成功、変更通知識別子を返却
	fmt.Printf("SUCCESS: UPDATE TWebHooks SET ExpireStamp: %d\n", nRows)

	pTx.Commit()

	return nRows, nil
}
