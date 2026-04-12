package googleapi

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// googleOauthConfig
var googleOauthConfig = &oauth2.Config{
	RedirectURL:  "YOUR_REDIRECT_URI", // 例: "http://localhost:8080/auth/google/callback"
	ClientID:     "YOUR_CLIENT_ID",
	ClientSecret: "YOUR_CLIENT_SECRET",
	Scopes: []string{
		"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/userinfo.profile",
		"https://www.googleapis.com/auth/calendar.events"},
	Endpoint: google.Endpoint,
}

// Google User Info APIからのレスポンスを格納する構造体
type UserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

// HandleGoogleLogin ログイン開始ハンドラ
func HandleGoogleLogin(w http.ResponseWriter, r *http.Request, pChallenge string) {
	googleOauthConfig.ClientID = os.Getenv("GOOGLE_CLIENT_ID")
	googleOauthConfig.ClientSecret = os.Getenv("GOOGLE_SECRET_KEY")
	googleOauthConfig.RedirectURL = "https://api.arteria-s.net:8443/auth/google/response"

	pURL := googleOauthConfig.AuthCodeURL(pChallenge, oauth2.AccessTypeOffline)
	log.Printf("HandleGoogleLogin(): %v\n", googleOauthConfig)

	http.Redirect(w, r, pURL, http.StatusTemporaryRedirect)
}

// HandleGoogleGrant アクセス許可申請を開始
func HandleGoogleGrant(w http.ResponseWriter, r *http.Request, pChallenge string) {
	googleOauthConfig.ClientID = os.Getenv("GOOGLE_CLIENT_ID")
	googleOauthConfig.ClientSecret = os.Getenv("GOOGLE_SECRET_KEY")
	googleOauthConfig.RedirectURL = "https://api.arteria-s.net:8443/perm/google/response"

	// oauth2.AccessTypeOfflineを指定するとリフレッシュトークンも発行される。
	pURL := googleOauthConfig.AuthCodeURL(pChallenge, oauth2.AccessTypeOffline)
	log.Printf("HandleGoogleGrant(): %v\n", googleOauthConfig)

	http.Redirect(w, r, pURL, http.StatusTemporaryRedirect)
}

// HandleGoogleRevoke アクセストークン破棄処理ハンドラ
func HandleGoogleRevoke(w http.ResponseWriter, r *http.Request, pAccessToken string) {
	// フォームデータを準備
	// Googleのrevokeエンドポイントは、トークンを「application/x-www-form-urlencoded」形式で受け取る
	formData := url.Values{}
	formData.Set("token", pAccessToken)

	// POSTリクエストを作成
	// strings.NewReader()でフォームデータを io.Reader として渡す
	pRevokeURL := "https://oauth2.googleapis.com/revoke"
	req, pError := http.NewRequest("POST", pRevokeURL, strings.NewReader(formData.Encode()))
	if pError != nil {
		_ = fmt.Errorf("リクエストの作成に失敗しました: %w", pError)
		return
	}

	// Content-Type ヘッダーを忘れずに設定
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// HTTPクライアントを作成し、リクエストを送信
	client := &http.Client{}
	resp, pError := client.Do(req)
	if pError != nil {
		_ = fmt.Errorf("HTTPリクエストの実行に失敗しました: %w", pError)
		return
	}
	defer resp.Body.Close()

	// 成功の場合、ステータスコードは 200 (OK) が返される
	if resp.StatusCode != http.StatusOK {
		// トークンが無効だった場合でも 200 が返されることがあるが、
		// 400などのエラーが返された場合は、通常、サーバー側の問題ではない
		_ = fmt.Errorf("トークン無効化が失敗しました: ステータスコード %d", resp.StatusCode)
		return
	}
}

// HandleGoogleCallback コールバックハンドラ
func HandleGoogleCallback(w http.ResponseWriter, r *http.Request, pAccessToken *oauth2.Token) error {
	// 認証コードを取得
	code := r.FormValue("code")

	// 認証コードとAccess Tokenを交換
	pToken, pError := googleOauthConfig.Exchange(r.Context(), code)
	if pError != nil {
		// エラー処理
		return fmt.Errorf("HandleGoogleCallback(): googleOauthConfig.Exchange() %s", pError.Error())
	}
	log.Printf("Google granted request. Access Token Expire at %s\n", pToken.Expiry.Format(time.RFC1123Z))

	*pAccessToken = *pToken

	return nil
}

// HandleGoogleRefreshToken アクセストークンを更新
func HandleGoogleRefreshToken(r *http.Request, fForce bool, pAccessToken *oauth2.Token) (*oauth2.Token, error) {
	googleOauthConfig.ClientID = os.Getenv("GOOGLE_CLIENT_ID")
	googleOauthConfig.ClientSecret = os.Getenv("GOOGLE_SECRET_KEY")
	googleOauthConfig.RedirectURL = "https://api.arteria-s.net:8443/perm/google/response"

	pExpireAccessToken := pAccessToken.AccessToken
	pExpireRefreshToken := pAccessToken.RefreshToken

	// アクセストークンを強制する場合は、トークンの有効期限を過去に置き換える。
	if fForce {
		pAccessToken.Expiry = time.Now().Add(time.Hour * -1)
		log.Printf("Tok: %s\n", pAccessToken.Expiry.Format(time.RFC1123Z))
		log.Printf("Now: %s\n", time.Now().Format(time.RFC1123Z))
	}
	if pAccessToken.Expiry.Before(time.Now()) {
		pTokenSource := googleOauthConfig.TokenSource(r.Context(), pAccessToken)
		pNewToken, pError := pTokenSource.Token()
		if pError != nil {
			return nil, pError
		}
		if pExpireAccessToken == pNewToken.AccessToken {
			log.Println("ERROR: unable to change Access Token.")
		}
		if pExpireRefreshToken == pNewToken.RefreshToken {
			log.Println("ERROR: unable to change Refresh Token.")
		}
		return pNewToken, nil
	}
	return pAccessToken, nil
}
