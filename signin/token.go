package signin

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"arteria-s.net/entraapi"
	"arteria-s.net/googleapi"
	"arteria-s.net/postgres"
	"golang.org/x/oauth2"
)

// pApplicationId アプリケーション識別子
//var pApplicationId = "dec981fb-da92-4dde-82a9-8b80bae80071"

// PrepareGoogleToken アクセストークンの有効性を確認し、必要であれば有効なトークンに更新する。
func PrepareGoogleToken(r *http.Request, pDatabase *sql.DB, pUniqueId string, fForce bool, pG *GoogleContext) error {
	pNewToken, pError := googleapi.HandleGoogleRefreshToken(r, fForce, &pG.pAToken)
	if pError != nil {
		return pError
	}
	pError = postgres.UpsertDelegateToken(pDatabase, postgres.AuthorityGoogle, pUniqueId, pNewToken)
	if pError != nil {
		log.Println("FAILED: PrepareGoogleToken().UpsertDelegateToken()")
	}
	log.Printf("SUCCESS: Prepared AccessToken of Google Id. Expired at %s\n", pNewToken.Expiry.Format(time.RFC1123Z))
	pG.pAToken = *pNewToken

	return nil
}

// PrepareEntraToken アクセストークンの有効性を確認し、必要であれば有効なトークンに更新する。
func PrepareEntraToken(r *http.Request, pDatabase *sql.DB, pUniqueId string, fForce bool, pE *EntraContext) error {
	pNewToken, pError := entraapi.HandleEntraRefreshToken(r, fForce, &pE.pAToken)
	if pError != nil {
		return pError
	}
	pError = postgres.UpsertDelegateToken(pDatabase, postgres.AuthorityEntra, pUniqueId, pNewToken)
	if pError != nil {
		log.Println("FAILED: PrepareEntraToken().UpsertDelegateToken()")
	}
	log.Printf("SUCCESS: Prepared AccessToken of Entra Id. Expired at %s\n", pNewToken.Expiry.Format(time.RFC1123Z))
	pE.pAToken = *pNewToken

	return nil
}

// SetSessionToken アクセストークン群を保存
func SetSessionToken(pDatabase *sql.DB, eAuthority int, pSessionKey string, pToken *oauth2.Token) error {
	pError := postgres.CreateSessionToken(pDatabase, eAuthority, pSessionKey, pToken)
	return pError
}

// GetSessionTokenByKey アクセストークン群を獲得
func GetSessionTokenByKey(pDatabase *sql.DB, eAuthority int, pSessionKey string, pToken *oauth2.Token) (int, error) {
	nRows, pError := postgres.LookupSessionTokenByKey(pDatabase, eAuthority, pSessionKey, pToken)
	return nRows, pError
}

// DeleteSessionToken アクセストークンを削除
func DeleteSessionToken(pDatabase *sql.DB, eAuthority int, pSessionKey string) error {
	pError := postgres.DeleteSessionToken(pDatabase, eAuthority, pSessionKey)
	return pError
}

// CheckEndpoint エンドポイントとセッションの正常性を検査
func CheckEndpoint(pDatabase *sql.DB, pSessionKey string, pUniqueId *string) (int, error) {
	// セッションキーから利用者識別子の記録を取得
	pError := postgres.GetUniqueIdFromSession(pDatabase, pSessionKey, pUniqueId)
	if pError != nil {
		return 0, pError
	}

	return 0, nil
}

// GetDelegateTokenByID アクセストークン群を獲得
func GetDelegateTokenById(pDatabase *sql.DB, eAuthority int, pUniqueId string, pToken *oauth2.Token) (int, error) {
	nRows, pError := postgres.LookupDelegateTokenById(pDatabase, eAuthority, pUniqueId, pToken)
	return nRows, pError
}

// DeleteDelegateToken アクセストークンを獲得
func DeleteDelegateToken(pDatabase *sql.DB, eAuthority int, pUniqueId string) error {
	pError := postgres.DeleteDelegateToken(pDatabase, eAuthority, pUniqueId)
	return pError
}

// SetDelegateToken アクセストークンを記録
func SetDelegateToken(pDatabase *sql.DB, eAuthority int, pUniqueId string, pToken *oauth2.Token) error {
	pError := postgres.UpsertDelegateToken(pDatabase, eAuthority, pUniqueId, pToken)
	return pError
}
