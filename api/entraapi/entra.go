package entraapi

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"golang.org/x/oauth2"
)

// SDKをダウンロードするコマンド
// go get github.com/microsoftgraph/msgraph-sdk-go
// go get github.com/Azure/azure-sdk-for-go/sdk/azidentity
// （2025/10/07 現在）
// "github.com/microsoftgraph/msgraph-sdk-go/models"をインポートするとビルドを行い始める。
// https://learn.microsoft.com/en-us/graph/sdks/generate-with-kiota
// → SDKを使わない方がよい。（2025/10/07）

// Google User Info APIからのレスポンスを格納する構造体

// EntraOauthConfig
var pEntraOauthConfig = &oauth2.Config{
	RedirectURL:  "YOUR_REDIRECT_URI",
	ClientID:     "YOUR_CLIENT_ID",
	ClientSecret: "YOUR_CLIENT_SECRET",
	Scopes:       []string{"openid", "profile", "offline_access", "User.Read", "Calendars.Read", "User.RevokeSessions.All"},
}

// Entra User Info APIからのレスポンスを格納する構造体
type UserInfo struct {
	ID         string `json:"id"`
	Email      string `json:"mail"`
	Name       string `json:"displayName"`
	GivenName  string `json:"givenName"`
	FamilyName string `json:"surname"`
	Picture    string `json:"picture"`
}

// RevokeResponse 失効応答
type RevokeResponse struct {
	Value bool `json:"Value"`
}

// HandleEntraLogin ログイン開始ハンドラ
func HandleEntraLogin(w http.ResponseWriter, r *http.Request, pChallenge string) {

	pEntraOauthConfig.ClientID = os.Getenv("ENTRA_CLIENT_ID")
	pEntraOauthConfig.ClientSecret = os.Getenv("ENTRA_SECRET_KEY")
	pEntraOauthConfig.RedirectURL = "https://api.arteria-s.net:8443/auth/entra/response"

	// テナント固有のエンドポイント
	pTenantID := os.Getenv("ENTRA_TENANT_ID")
	pEntraOauthConfig.Endpoint = oauth2.Endpoint{
		AuthURL:  fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/authorize", pTenantID),
		TokenURL: fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", pTenantID),
	}

	pOptions := oauth2.SetAuthURLParam("prompt", "select_account")
	pURL := pEntraOauthConfig.AuthCodeURL(pChallenge, oauth2.AccessTypeOnline, pOptions)

	http.Redirect(w, r, pURL, http.StatusTemporaryRedirect)
}

// HandleEntraGrant アクセス許可申請を開始
func HandleEntraGrant(w http.ResponseWriter, r *http.Request, pChallenge string) {

	pEntraOauthConfig.ClientID = os.Getenv("ENTRA_CLIENT_ID")
	pEntraOauthConfig.ClientSecret = os.Getenv("ENTRA_SECRET_KEY")
	pEntraOauthConfig.RedirectURL = "https://api.arteria-s.net:8443/perm/entra/response"

	// テナント固有のエンドポイント
	pTenantID := os.Getenv("ENTRA_TENANT_ID")
	pEntraOauthConfig.Endpoint = oauth2.Endpoint{
		AuthURL:  fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/authorize", pTenantID),
		TokenURL: fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", pTenantID),
	}

	// アカウント選択を促す。
	pOptions := oauth2.SetAuthURLParam("prompt", "select_account")
	pURL := pEntraOauthConfig.AuthCodeURL(pChallenge, oauth2.AccessTypeOnline, pOptions)

	http.Redirect(w, r, pURL, http.StatusTemporaryRedirect)
}

// HandleEntraLogout ログアウト処理ハンドラ
func HandleEntraLogout(w http.ResponseWriter, r *http.Request, pAccessToken string) {
	// フォームデータを準備
	// Entraのrevokeエンドポイントは、トークンを「application/x-www-form-urlencoded」形式で受け取る
	formData := url.Values{}
	formData.Set("token", pAccessToken)
	formData.Add("post_logout_redirect_uri", "https://api.arteria-s.net:8443/auth/logout")
	//formData.Add("id_token_hint", pAccessToken)

	pTenantID := os.Getenv("ENTRA_TENANT_ID")
	pRevokeURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/logout", pTenantID)
	pLogoutURL := pRevokeURL + "?" + formData.Encode()

	// ログアウトURLにリダイレクト
	http.Redirect(w, r, pLogoutURL, http.StatusTemporaryRedirect)
}

// HandleEntraCallback コールバックハンドラ
func HandleEntraCallback(w http.ResponseWriter, r *http.Request, pAccessToken *oauth2.Token) error {
	// 認証コードを取得
	code := r.FormValue("code")

	// 認証コードとAccess Tokenを交換
	pToken, pError := pEntraOauthConfig.Exchange(r.Context(), code)
	if pError != nil {
		// エラー処理
		return fmt.Errorf("HandleEntraCallback(): pEntraOauthConfig.Exchange() %s", pError.Error())
	}
	log.Printf("Entra granted request. Access Token Expire at %s\n", pToken.Expiry.Format(time.RFC1123Z))
	*pAccessToken = *pToken

	return nil
}

// HandleEntraRevoke アクセストークン破棄処理ハンドラ
func HandleEntraRevoke(r *http.Request, pAccessToken *oauth2.Token, pValue *bool) {
	// Access Tokenを使用してアクセストークンの破棄を実施
	pEndPoint := "https://graph.microsoft.com/v1.0/me/revokeSignInSessions"
	pRequest, pError := http.NewRequest("POST", pEndPoint, nil)
	if pError != nil {
		log.Printf("http.NewRequest(): %s", pError.Error())
		return
	}
	pRequest.Header.Set("Authorization", "Bearer "+pAccessToken.AccessToken)
	pRequest.Header.Set("Accept", "application/json")

	pClient := &http.Client{Timeout: 10 * time.Second}
	pResponse, pError := pClient.Do(pRequest)
	if pError != nil {
		log.Printf("http.Client().Do(): %s", pError.Error())
		return
	}
	defer pResponse.Body.Close()

	if pResponse.StatusCode != http.StatusOK {
		// エラーレスポンスのボディを読み取って詳細を出力できます
		bodyBytes, _ := io.ReadAll(pResponse.Body)
		log.Printf("HandleEntraRevoke(): %s", string(bodyBytes))
		return
	}

	var pRevokeResponse RevokeResponse
	if pError := json.NewDecoder(pResponse.Body).Decode(&pRevokeResponse); pError != nil {
		log.Printf("json.NewDecoder(): %s", pError.Error())
		return
	}

	*pValue = pRevokeResponse.Value
}

// HandleEntraRefreshToken アクセストークンを更新
func HandleEntraRefreshToken(r *http.Request, fForce bool, pAccessToken *oauth2.Token) (*oauth2.Token, error) {
	pEntraOauthConfig.ClientID = os.Getenv("ENTRA_CLIENT_ID")
	pEntraOauthConfig.ClientSecret = os.Getenv("ENTRA_SECRET_KEY")
	pEntraOauthConfig.RedirectURL = "https://api.arteria-s.net:8443/perm/entra/response"

	// テナント固有のエンドポイント
	pTenantID := os.Getenv("ENTRA_TENANT_ID")
	pEntraOauthConfig.Endpoint = oauth2.Endpoint{
		AuthURL:  fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/authorize", pTenantID),
		TokenURL: fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", pTenantID),
	}

	pExpireAccessToken := pAccessToken.AccessToken
	pExpireRefreshToken := pAccessToken.RefreshToken

	// アクセストークンを強制する場合は、トークンの有効期限を過去に置き換える。
	if fForce {
		pAccessToken.Expiry = time.Now().Add(time.Hour * -1)
		log.Printf("Tok: %s\n", pAccessToken.Expiry.Format(time.RFC1123Z))
		log.Printf("Now: %s\n", time.Now().Format(time.RFC1123Z))
	}
	if pAccessToken.Expiry.Before(time.Now()) {
		pTokenSource := pEntraOauthConfig.TokenSource(r.Context(), pAccessToken)
		pNewToken, pError := pTokenSource.Token()
		if pError != nil {
			log.Println("ERROR: Failure TokenSource.Token().")
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
