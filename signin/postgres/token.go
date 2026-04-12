package postgres

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"golang.org/x/oauth2"
)

// AuthStatusPermmit 許可
const AuthStatusPermmit = 100

// AuthStatusAccept 受入
const AuthStatusAccept = 101

// AuthStatusFailed 認証失敗
const AuthStatusFailed = 1

// AuthStatusDenied 拒否
const AuthStatusDenied = 2

// AuthStatusReject 却下
const AuthStatusReject = 3

const (
	AuthorityGoogle = iota
	AuthorityEntra
)

// Authoritys 認証サーバー識別子配列
var Authoritys = []string{"68d14f7b-b5ee-401e-b338-15efad4f87c2", "7bba2244-a60c-445e-9c8a-ab4cdf4e2473"}
var AuthorityName = []string{"Google", "Entra"}

// CreateSessionToken セッション情報を生成
func CreateSessionToken(pDatabase *sql.DB, eAuthority int, pSessionKey string, pToken *oauth2.Token) error {
	// トランザクションを開始
	pTx, pError := pDatabase.Begin()
	if pError != nil {
		// トランザクションの開始に失敗
		fmt.Println("Failed: Begin(): " + pError.Error())
		return pError
	}

	// アクセストークンの記録
	pExpireStamp := pToken.Expiry.Format(time.RFC1123Z)
	const pSQL = "INSERT INTO TSTokens (AuthorityId, SessionKey, AccessToken, RefreshToken, Expiry) VALUES ($1, $2, $3, $4, $5);"
	pResult, pError := pTx.Exec(pSQL, Authoritys[eAuthority], pSessionKey, pToken.AccessToken, pToken.RefreshToken, pExpireStamp)
	if pError != nil {
		fmt.Printf("FAILED: CreateToken(): AuthorityId=[%s], SessionKey=[%s], A-Token[%s], R-Token=[%s], Expire=[%s]\n", Authoritys[eAuthority], pSessionKey, pToken.AccessToken, pToken.RefreshToken, pExpireStamp)
		fmt.Printf("INSERT INTO TSTokens (AuthorityId, SessionKey, AccessToken, RefreshToken, Expiry) VALUES ('%s', '%s', '%s', '%s', '%s');\n", Authoritys[eAuthority], pSessionKey, pToken.AccessToken, pToken.RefreshToken, pExpireStamp)

		if pPQError, ok := pError.(*pq.Error); ok {
			pSQLError := pPQError.Code.Name()
			if strings.Compare(pSQLError, "unique_violation") == 0 {
				//　既にアクセストークンの記録されている
				fmt.Println("CreateSessionToken(): 既にアクセストークンが記録されている。 SessionKey=[" + pSessionKey + "]")
				fmt.Println(pError)
				pTx.Rollback()
				return pError
			} else if strings.Compare(pSQLError, "invalid_authorization_specification") == 0 {
				fmt.Println("データベースインスタンスからアクセスを拒否されました。")
				fmt.Println(pError)
				pTx.Rollback()
				return pError
			} else {
				fmt.Println("CreateSessionToken()::pSQLError: [" + pSQLError + "]")
			}
		}
		// アクセストークンの記録に失敗
		fmt.Println("アクセストークンの記録に失敗しました。[" + pSessionKey + "]")
		fmt.Println(pError)
		pTx.Rollback()
		return pError
	}

	nRows, pError := pResult.RowsAffected()
	if pError != nil {
		if pPQError, ok := pError.(*pq.Error); ok {
			pSQLError := pPQError.Code.Name()
			fmt.Println("pSQLError: [" + pSQLError + "]")
		}
		pTx.Rollback()
		return pError
	}

	if nRows == 0 {
		fmt.Println("アクセストークンの記録に失敗しました。処理行数が0件です。[" + pSessionKey + "]")
		pTx.Rollback()
		return fmt.Errorf("CreateSessionToken() count=0")
	}

	// アクセストークンの記録に成功、セッション識別子を返却
	pTx.Commit()

	return nil
}

// DeleteSessionToken セッション情報を生成
func DeleteSessionToken(pDatabase *sql.DB, eAuthority int, pSessionKey string) error {
	if pDatabase == nil {
		return fmt.Errorf("DeleteSessionToken(): pDatabase")
	}
	pTransaction, pError := Begin(pDatabase)
	if pError != nil {
		return fmt.Errorf("DeleteSessionToken(): Begin()")
	}
	const pSQL = "DELETE FROM TSTokens WHERE SessionKey = $1 AND AuthorityId = $2;"
	if !(pTransaction.Execute(pSQL, pSessionKey, Authoritys[eAuthority])) {
		pTransaction.Rollback()
		return fmt.Errorf("DeleteSessionToken(): Execute()")
	}

	pTransaction.Commit()

	return nil
}

// DeleteSessionTokenByToken セッション情報を削除
func DeleteSessionTokenByToken(pDatabase *sql.DB, eAuthority int, pSessionKey string, pToken string) error {
	if pDatabase == nil {
		return fmt.Errorf("DeleteSessionTokenByToken(): pDatabase")
	}
	pTransaction, pError := Begin(pDatabase)
	if pError != nil {
		return fmt.Errorf("DeleteSessionTokenByToken(): Begin()")
	}
	const pSQL = "DELETE FROM TSTokens WHERE SessionKey = $1 AND AuthorityId = $2 AND AccessToken = $3;"
	if !(pTransaction.Execute(pSQL, pSessionKey, Authoritys[eAuthority], pToken)) {
		pTransaction.Rollback()
		return fmt.Errorf("DeleteSessionTokenByToken(): Execute()")
	}

	pTransaction.Commit()

	return nil
}

// LookupSessionTokenByKey セッション情報からセッション識別子に対応するレコードを検索
func LookupSessionTokenByKey(pDatabase *sql.DB, eAuthority int, pSessionKey string, pToken *oauth2.Token) (int, error) {
	if pDatabase == nil {
		return 0, fmt.Errorf("LookupSessionTokenByKey(): Invalid parameter at Database. eAuthority=[%d], SessionKey=[%s]", eAuthority, pSessionKey)
	}
	pTransaction, pError := Begin(pDatabase)
	if pError != nil {
		return 0, fmt.Errorf("LookupSessionTokenByKey(): Failure transaction begin. eAuthority=[%d], SessionKey=[%s]", eAuthority, pSessionKey)
	}
	defer pTransaction.Commit()

	const pSQL = "SELECT AccessToken, RefreshToken, Expiry FROM TSTokens WHERE AuthorityId = $1 AND SessionKey = $2;"
	//log.Printf("SELECT AccessToken, RefreshToken, Expiry FROM TSTokens WHERE AuthorityId = '%s' AND SessionKey = '%s';", Authoritys[eAuthority], pSessionKey)
	pAuthorityId := Authoritys[eAuthority]
	pRows, pErr := pTransaction.Query(pSQL, pAuthorityId, pSessionKey)
	if pErr != nil {
		pTransaction.Rollback()
		return 0, fmt.Errorf("LookupSessionTokenByKey(): Failure SELECT Query.SQL=[%s],eAuthority=[%d], SessionKey=[%s]", pSQL, eAuthority, pSessionKey)
	}
	defer pRows.Close()

	var nRows int = 0
	for pRows.Next() {
		if pErr = pRows.Scan(&pToken.AccessToken, &pToken.RefreshToken, &pToken.Expiry); pErr != nil {
			if pPQError, ok := pErr.(*pq.Error); ok {
				pSQLError := pPQError.Code.Name()
				if strings.Compare(pSQLError, "invalid_authorization_specification") == 0 {
					return 0, fmt.Errorf("LookupSessionTokenByKey(): データベースインスタンスからアクセスを拒否されました。SessionKey=[%s],SQLError=[%s]", pSessionKey, pSQLError)
				}
				return 0, fmt.Errorf("LookupSessionTokenByKey(): アクセストークンの入力に失敗しました。SessionKey=[%s],SQLError=[%s]", pSessionKey, pSQLError)
			}
			return 0, fmt.Errorf("LookupSessionTokenByKey(): アクセストークンの入力に失敗しました。eAuthority=[%d], SessionKey=[%s]", eAuthority, pSessionKey)
		} else {
			nRows++
			break
		}
	}

	return nRows, nil
}

// LookupDelegateTokenById セッション情報からセッション識別子に対応するレコードを検索
func LookupDelegateTokenById(pDatabase *sql.DB, eAuthority int, pUniqueId string, pToken *oauth2.Token) (int, error) {
	if pDatabase == nil {
		return 0, fmt.Errorf("LookupDelegateTokenById(): Invalid parameter at Database. eAuthority=[%d], UniqueId=[%s]", eAuthority, pUniqueId)
	}
	pTransaction, pError := Begin(pDatabase)
	if pError != nil {
		return 0, fmt.Errorf("LookupDelegateTokenById(): Failure transaction begin. eAuthority=[%d], UniqueId=[%s]", eAuthority, pUniqueId)
	}
	defer pTransaction.Commit()

	const pSQL = "SELECT AccessToken, RefreshToken, Expiry FROM TATokens WHERE AuthorityId = $1 AND UniqueId = $2;"
	pAuthorityId := Authoritys[eAuthority]
	pRows, pErr := pTransaction.Query(pSQL, pAuthorityId, pUniqueId)
	if pErr != nil {
		pTransaction.Rollback()
		return 0, fmt.Errorf("LookupDelegateTokenById(): Failure SELECT Query.SQL=[%s], eAuthority=[%d], UniqueId=[%s]", pSQL, eAuthority, pUniqueId)
	}
	defer pRows.Close()

	var nRows int = 0
	for pRows.Next() {
		if pErr = pRows.Scan(&pToken.AccessToken, &pToken.RefreshToken, &pToken.Expiry); pErr != nil {
			if pPQError, ok := pErr.(*pq.Error); ok {
				pSQLError := pPQError.Code.Name()
				if strings.Compare(pSQLError, "invalid_authorization_specification") == 0 {
					return 0, fmt.Errorf("LookupDelegateTokenById(): データベースインスタンスからアクセスを拒否されました。pUniqueId=[%s], SQLError=[%s]", pUniqueId, pSQLError)
				}
				return 0, fmt.Errorf("LookupDelegateTokenById(): アクセストークンの入力に失敗しました。pUniqueId=[%s], SQLError=[%s]", pUniqueId, pSQLError)
			}
			return 0, fmt.Errorf("LookupDelegateTokenById(): アクセストークンの入力に失敗しました。eAuthority=[%d], UniqueId=[%s]", eAuthority, pUniqueId)
		} else {
			nRows++
			break
		}
	}

	return nRows, nil
}

// DeleteDelegateToken アクセストークンを削除
func DeleteDelegateToken(pDatabase *sql.DB, eAuthority int, pUniqueId string) error {
	if pDatabase == nil {
		return fmt.Errorf("DeleteDelegateToken(): pDatabase")
	}
	pTransaction, pError := Begin(pDatabase)
	if pError != nil {
		return fmt.Errorf("DeleteDelegateToken(): Begin()")
	}
	const pSQL = "DELETE FROM TATokens WHERE UniqueId = $1 AND AuthorityId = $2;"
	if !(pTransaction.Execute(pSQL, pUniqueId, Authoritys[eAuthority])) {
		pTransaction.Rollback()
		return fmt.Errorf("DeleteDelegateToken(): Execute()")
	}

	pTransaction.Commit()

	return nil
}

// UpsertDelegateToken 代理用アクセストークンを記録
func UpsertDelegateToken(pDatabase *sql.DB, eAuthority int, pUniqueId string, pToken *oauth2.Token) error {
	// トランザクションを開始
	pTx, pError := pDatabase.Begin()
	if pError != nil {
		// トランザクションの開始に失敗
		fmt.Println("Failed: Begin(): " + pError.Error())
		return pError
	}

	// アクセストークンの記録
	pExpireStamp := pToken.Expiry.Format(time.RFC1123Z)
	const pSQL = "INSERT INTO TATokens (AuthorityId, UniqueId, AccessToken, RefreshToken, Expiry) VALUES ($1, $2, $3, $4, $5) ON CONFLICT ON CONSTRAINT tatokens_pkey DO UPDATE SET AccessToken = $3, RefreshToken = $4, Expiry = $5;"
	pResult, pError := pTx.Exec(pSQL, Authoritys[eAuthority], pUniqueId, pToken.AccessToken, pToken.RefreshToken, pExpireStamp)
	if pError != nil {
		fmt.Printf("FAILED: UpsertDelegateToken(): AuthorityId=[%s], UniqueId=[%s], A-Token[%s], R-Token=[%s], Expire=[%s]\n", Authoritys[eAuthority], pUniqueId, pToken.AccessToken, pToken.RefreshToken, pExpireStamp)
		fmt.Printf("INSERT INTO TATokens (AuthorityId, UniqueId, AccessToken, RefreshToken, Expiry) VALUES ('%s', '%s', '%s', '%s', '%s');\n", Authoritys[eAuthority], pUniqueId, pToken.AccessToken, pToken.RefreshToken, pExpireStamp)

		if pPQError, ok := pError.(*pq.Error); ok {
			pSQLError := pPQError.Code.Name()
			if strings.Compare(pSQLError, "unique_violation") == 0 {
				//　既にアクセストークンの記録されている
				fmt.Println("UpsertDelegateToken(): 既にアクセストークンが記録されている。 UniqueId=[" + pUniqueId + "]")
				fmt.Println(pError)
			} else if strings.Compare(pSQLError, "invalid_authorization_specification") == 0 {
				fmt.Println("UpsertDelegateToken(): データベースインスタンスからアクセスを拒否されました。")
				fmt.Println(pError)
			} else {
				fmt.Println("UpsertDelegateToken()::pSQLError: [" + pSQLError + "]")
			}
		}
		// アクセストークンの記録に失敗
		fmt.Println("UpsertDelegateToken(): アクセストークンの記録に失敗しました。[" + pUniqueId + "]")
		fmt.Println(pError)
		pTx.Rollback()
		return pError
	}

	nRows, pError := pResult.RowsAffected()
	if pError != nil {
		if pPQError, ok := pError.(*pq.Error); ok {
			pSQLError := pPQError.Code.Name()
			fmt.Println("pSQLError: [" + pSQLError + "]")
		}
		pTx.Rollback()
		return pError
	}

	if nRows == 0 {
		fmt.Println("アクセストークンの記録に失敗しました。処理行数が0件です。[" + pUniqueId + "]")
		pTx.Rollback()
		return fmt.Errorf("UpsertDelegateToken() count=0")
	}

	// アクセストークンの記録に成功、セッション識別子を返却
	pTx.Commit()

	return nil
}
