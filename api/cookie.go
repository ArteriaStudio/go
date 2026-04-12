package api

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"
)

// PickupCookie クッキーから指定項目の値を入力
func PickupCookie(r *http.Request, pKey string) string {
	//　セッション識別子をクッキーから獲得
	pCookie, pErr := r.Cookie(pKey)
	if pErr != nil {
		return ""
	}
	return pCookie.Value
}

// SetCookie クッキーを出力
func SetCookie(w http.ResponseWriter, pKey string, pToken string, AssociateDomain string, iMaxAge int) {
	var cookie http.Cookie
	cookie.Name = pKey
	cookie.Value = pToken
	cookie.Path = "/"
	cookie.Domain = AssociateDomain
	cookie.Expires = time.Now().AddDate(0, 0, 1)
	cookie.MaxAge = iMaxAge
	cookie.Secure = true
	cookie.HttpOnly = true
	//cookie.SameSite = http.SameSiteNoneMode
	//cookie.SameSite = http.SameSiteStrictMode	//　このモードの場合は、ブラウザはクッキーを必ず送信しない。ステートレスなアプリケーションが使う？そもステートレスならクッキーを送る必要なくね？
	cookie.SameSite = http.SameSiteLaxMode
	//cookie.SameSite = http.SameSiteDefaultMode
	http.SetCookie(w, &cookie)
}

// セキュリティのためのstate文字列を生成
func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
