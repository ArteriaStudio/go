// 　サインイン処理
package api

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"arteria-s.net/entraapi"
	"arteria-s.net/googleapi"
	"arteria-s.net/postgres"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"golang.org/x/oauth2"
)

// 依存関係を増減させた時に実行するコマンド（2025/10/07）
// go mod tidy; go mod download; go get -u
// 特に「go get -u」はコンパイル済みバイナリをダウンロードするのだが、これがないと依存モジュールをソースコードからビルドすることがある。

// pAssociateDomain 応答に乗せるドメイン名
var pAssociateDomain = "api.arteria-s.net"

// pApplicationName アプリケーション名
var pApplicationName = "api.arteria-s.net"

const (
	TokenGoogle = iota // Google
	TokenEntra         // Entra
)

// GoogleContext コンテキスト（Google）
type GoogleContext struct {
	pSToken   oauth2.Token       // アクセストークン（セッション）
	pAToken   oauth2.Token       // アクセストークン（代理）
	pId       string             // ユーザー識別子
	pUserInfo googleapi.UserInfo // ユーザー情報
}

// EntraContext コンテキスト（Entra）
type EntraContext struct {
	pSToken   oauth2.Token      // アクセストークン（セッション）
	pAToken   oauth2.Token      // アクセストークン（代理）
	pId       string            // ユーザー識別子
	pUserInfo entraapi.UserInfo // ユーザー情報
}

// FunctionContext 関数コンテキスト
type FunctionContext struct {
	pUniqueId    string        // 利用者ユニークキー
	pDisplayName string        // 表示名
	pSessionKey  string        // セッションキー
	pClientIP    string        // エンドポイントのグローバルIPアドレス
	pG           GoogleContext // コンテキスト（Google）
	pE           EntraContext  // コンテキスト（Entra）
	pDatabase    *sql.DB       // データベース接続
}

// システムパラメータ（Params）
var GParams Params

// init インスタンスを初期化
func init() {
	functions.HTTP("EntryPoint", EntryPoint)
	GParams.Initialize()
}

// EntryPoint エントリーポイント
func EntryPoint(w http.ResponseWriter, r *http.Request) {
	//　リクエストをフィルタ
	if CheckRequest(r) {
		//　応答せずに終了
		pClientIP := GetClientIP(r)
		log.Println("ClientIP: " + pClientIP)
		return
	}

	// WellKnown URIを処理
	if IsCheckWellKnownURI(r.URL.Path) {
		HandlerWellKnownURI(w, r, r.URL.Path)
		return
	}

	// Listener イベントを処理
	if IsListenerURI(r.URL.Path) {
		HandlerListenerURI(w, r, r.URL.Path)
	} else {
		HandlerInteractURI(w, r, r.URL.Path)
	}
}

// HandlerInteractURI
func HandlerInteractURI(w http.ResponseWriter, r *http.Request, pURI string) {
	// 関数コンテキストを作成
	var c FunctionContext
	var pCalendar Calendar

	// データベースと接続
	pServerHostName := "localhost"
	pDatabaseName := "abook"
	c.pDatabase = postgres.Open(pServerHostName, pDatabaseName)
	defer postgres.Close(c.pDatabase)

	// エンドポイントのグローバルIPアドレスを獲得
	c.pClientIP = GetClientIP(r)

	// セッション識別子を獲得 ※ クライアントセッションとリスナーセッションでは、アクセストークンへの紐付け方が異なる（2025/10/07）
	pError := PrepareSessionKey(r, &c)
	if pError != nil {
		log.Println("PrepareSessionKey(): " + pError.Error())
	}
	SetCookie(w, "XSRS-TOKEN", c.pSessionKey, pAssociateDomain, 86400)

	// セッション識別子にアクセストークンを紐付
	PrepareAccessTokens(r, &c)

	// イベントハンドラを実行
	switch r.URL.Path {
	case "/echo":
		HandlerEcho(w, r, &c)
	case "/page/signup":
		HandlerPageSignup(w, r, &c)
	case "/auth/google/signin":
		HandlerGoogleSignin(w, r, &c)
	case "/auth/google/response":
		HandlerGoogleAuthResponse(w, r, &c)
	case "/auth/entra/signin":
		HandlerEntraSignin(w, r, &c)
	case "/auth/entra/response":
		HandlerEntraAuthResponse(w, r, &c)
	case "/perm/google/grant":
		HandlerGoogleGrant(w, r, &c)
	case "/perm/google/revoke":
		HandlerGoogleRevoke(w, r, &c)
	case "/perm/google/response":
		HandlerGooglePermResponse(w, r, &c)
	case "/perm/entra/grant":
		HandlerEntraGrant(w, r, &c)
	case "/perm/entra/revoke":
		HandlerEntraRevoke(w, r, &c)
	case "/perm/entra/response":
		HandlerEntraPermResponse(w, r, &c)
	case "/profile/g":
		HandlerGooglePicture(w, r, &c)
	case "/profile/e":
		HandlerEntraPicture(w, r, &c)
	case "/subscribe/google":
		HandlerGoogleSubscribe(w, r, &c)
	case "/subscribe/entra":
		HandlerEntraSubscribe(w, r, &c)
	case "/subscribe/check":
		HandlerCheckSubscribe(w, r, &c)
	case "/unsubscribe/all":
		HandlerUnsubscribe(w, r, &c)
	case "/listup/google/subscribe":
		HandlerListupGoogleSubscribe(w, r, &c)
	case "/listup/entra/subscribe":
		HandlerListupEntraSubscribe(w, r, &c)
	case "/listup/google/calendar":
		pCalendar.HandlerListupGoogleCalendar(w, r, &c)
	case "/listup/entra/calendar":
		pCalendar.HandlerListupEntraCalendar(w, r, &c)
	case "/duplicate/google/calendar":
		pCalendar.HandlerDuplicateGoogleCalendar(w, r, &c)
	case "/duplicate/entra/calendar":
		pCalendar.HandlerDuplicateEntraCalendar(w, r, &c)
	case "/auth/signin":
		HandlerSignin(w, r, &c)
	case "/auth/signup":
		HandlerSignup(w, r, &c)
	case "/auth/signout":
		HandlerSignout(w, r, &c)
	case "/auth/logout":
		HandlerEntraLogout(w, r)
	default:
		if c.pUniqueId != "" {
			HandlerHome(w, r, &c)
		} else {
			HandlerDefault(w, r, &c)
		}
	}
}

// HandlerEcho エコー処理
func HandlerEcho(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	WriteHtmlHeaders(w, pApplicationName, http.StatusOK)

	fmt.Fprintln(w, "<body>")
	fmt.Fprintln(w, "<h2>エコー処理（arteria-s.net/echo）</h2>")

	DumpRequestHeader(w, r)
	DumpRequestParameters(w, r)
	DumpRequestBody(w, r)
	DumpClientIP(w, r)
	DumpSessionKey(w, r, c.pSessionKey)

	fmt.Fprintln(w, "[Google]<br />")
	fmt.Fprintln(w, "AccessToken: "+c.pG.pSToken.AccessToken+"<br />")
	fmt.Fprintln(w, "RefreshToken: "+c.pG.pSToken.RefreshToken+"<br />")
	fmt.Fprintln(w, "TokenType: "+c.pG.pSToken.TokenType+"<br />")

	fmt.Fprintln(w, "[Entra]<br />")
	fmt.Fprintln(w, "AccessToken: "+c.pE.pSToken.AccessToken+"<br />")
	fmt.Fprintln(w, "RefreshToken: "+c.pE.pSToken.RefreshToken+"<br />")
	fmt.Fprintln(w, "TokenType: "+c.pE.pSToken.TokenType+"<br />")

	fmt.Fprintln(w, "</body>")
}

// HandlerDefault 既定処理
func HandlerDefault(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	// 未ログイン
	WriteHtmlHeaders(w, pApplicationName, http.StatusOK)

	fmt.Fprintln(w, "<body>")
	fmt.Fprintln(w, "<h2>Portal</h2><hr /><p>")
	fmt.Fprintln(w, `<div class="login-container">`)
	fmt.Fprintln(w, `<form action="/auth/signin" method="post">`)
	fmt.Fprintln(w, `<div class="input-group">`)
	fmt.Fprintln(w, `<label for="email">メールアドレス:</label>`)
	fmt.Fprintln(w, `<input type="email" id="email" name="email" required placeholder="your.email@example.com">`)
	fmt.Fprintln(w, `</div>`)
	fmt.Fprintln(w, `<div class="input-group">`)
	fmt.Fprintln(w, `<label for="password">パスワード:</label>`)
	fmt.Fprintln(w, `<input type="password" id="password" name="password" required placeholder="********">`)
	fmt.Fprintln(w, `</div>`)
	fmt.Fprintln(w, `<button type="submit">ログイン</button>`)
	fmt.Fprintln(w, `</form>`)
	fmt.Fprintln(w, `</div>`)
	fmt.Fprintln(w, "<a href='/page/signup'>サインアップ</a>")
	fmt.Fprintln(w, "</p><hr />")
	fmt.Fprintln(w, "</body>")
}

// HandlerHome ログイン済のセッション向けの初期画面を出力
func HandlerHome(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	// ページに出力する情報を収集
	pError := PrepareHome(r, c)
	if pError != nil {
		pURL := "/"
		http.Redirect(w, r, pURL, http.StatusTemporaryRedirect)
	} else {
		// ページを出力
		WriteHtmlHeaders(w, pApplicationName, http.StatusOK)
		WriteHome(w, r, c)
	}
}

// PrepareHome ページ表示に必要な情報を収集
func PrepareHome(r *http.Request, c *FunctionContext) error {
	if c.pG.pAToken.AccessToken != "" {
		pError := PrepareGoogleToken(r, c.pDatabase, c.pUniqueId, false, &c.pG)
		if pError != nil {
			log.Printf("ERROR: PrepareGoogleToken: %s\n", pError.Error())
		}
		pError = googleapi.GetUserInfoGoogle(r, &c.pG.pAToken, &c.pG.pUserInfo)
		if pError != nil {
			// アクセストークンを破棄
			postgres.DeleteDelegateToken(c.pDatabase, postgres.AuthorityGoogle, c.pUniqueId)
			log.Println("ERROR: GetUserInfoGoogle(): " + pError.Error())
			return pError
		}
	}
	if c.pE.pAToken.AccessToken != "" {
		pError := PrepareEntraToken(r, c.pDatabase, c.pUniqueId, false, &c.pE)
		if pError != nil {
			log.Printf("ERROR: PrepareEntraToken: %s\n", pError.Error())
		}
		pError = entraapi.GetUserInfoEntra(r, &c.pE.pAToken, &c.pE.pUserInfo)
		if pError != nil {
			// アクセストークンを破棄
			postgres.DeleteDelegateToken(c.pDatabase, postgres.AuthorityEntra, c.pUniqueId)
			log.Println("ERROR: GetUserInfoEntra(): " + pError.Error())
			return pError
		}
	}
	return nil
}

// HandlerWellKnownURI 指定されたURIに対応するファイルを返送
func HandlerWellKnownURI(w http.ResponseWriter, r *http.Request, pURI string) {
	pFilepath := "assets" + pURI
	http.ServeFile(w, r, pFilepath)
}

// HandlerSignout セッション情報を削除
func HandlerSignout(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	// セッションを削除
	pError := postgres.DeleteSessionKey(c.pDatabase, c.pSessionKey)
	if pError != nil {
		log.Printf("ERROR: HandlerSignout().DeleteSessionKey() SessionKey=%s\n", c.pSessionKey)
	}

	//　クッキーを削除
	SetCookie(w, "XSRS-TOKEN", "", pAssociateDomain, -1)

	//　トップページにリダイレクト
	pURL := "/auth/logout"
	http.Redirect(w, r, pURL, http.StatusTemporaryRedirect)
}

// HandlerEntraLogout
func HandlerEntraLogout(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "<p>サインアウトしました。</p>")
	fmt.Fprintln(w, `<p><a href="/">ホームに戻る。</a></p>`)
}

// PrepareAccessTokens セッションキーにアクセストークンを紐付け
func PrepareAccessTokens(r *http.Request, c *FunctionContext) {
	// エンドポイントの正当性／妥当性を検査
	nRows, pError := CheckEndpoint(c.pDatabase, c.pSessionKey, &c.pUniqueId)
	if pError != nil {
		log.Println("PrepareAccessTokens()::CheckEndpoint(): " + pError.Error())
		return
	}
	if nRows != 0 {
		// セッションに異常を検出（アクセストークンを利用しない）
		return
	}

	pError = PrepareSessionAccessToken(c.pDatabase, postgres.AuthorityGoogle, c.pSessionKey, &c.pG.pSToken)
	if pError != nil {
		log.Println("PrepareAccessTokens()::PrepareSessionAccessToken(G): " + pError.Error())
	}
	pError = PrepareSessionAccessToken(c.pDatabase, postgres.AuthorityEntra, c.pSessionKey, &c.pE.pSToken)
	if pError != nil {
		log.Println("PrepareAccessTokens()::PrepareSessionAccessToken(E): " + pError.Error())
	}

	if c.pUniqueId != "" {
		pError = PrepareDelegateAccessToken(c.pDatabase, c.pUniqueId, postgres.AuthorityGoogle, &c.pG.pAToken)
		if pError != nil {
			log.Println("PrepareAccessTokens()::PrepareDelegateAccessToken(G): " + pError.Error())
		}
		if c.pG.pAToken.RefreshToken != "" {
			pNewToken, pError := googleapi.HandleGoogleRefreshToken(r, false, &c.pG.pAToken)
			if pError != nil {
				log.Println("PrepareAccessTokens()::HandleGoogleRefreshToken(G): " + pError.Error())
			} else {
				c.pG.pAToken = *pNewToken
			}
		}

		pError = PrepareDelegateAccessToken(c.pDatabase, c.pUniqueId, postgres.AuthorityEntra, &c.pE.pAToken)
		if pError != nil {
			log.Println("PrepareAccessTokens()::PrepareDelegateAccessToken(E): " + pError.Error())
		}
		if c.pE.pAToken.RefreshToken != "" {
			pNewToken, pError := entraapi.HandleEntraRefreshToken(r, false, &c.pE.pAToken)
			if pError != nil {
				log.Println("PrepareAccessTokens()::HandleEntraRefreshToken(E): " + pError.Error())
			} else {
				c.pE.pAToken = *pNewToken
			}
		}

		pError = GetAccountId(c.pDatabase, c.pUniqueId, postgres.AuthorityGoogle, &c.pG.pId)
		if pError != nil {
			log.Println("PrepareAccessTokens()::GetAccountId(G): " + pError.Error())
		}
		pError = GetAccountId(c.pDatabase, c.pUniqueId, postgres.AuthorityEntra, &c.pE.pId)
		if pError != nil {
			log.Println("PrepareAccessTokens()::GetAccountId(E): " + pError.Error())
		}
	}

	pError = IsExistsHistory(c.pDatabase, postgres.AuthorityGoogle, c.pG.pId, c.pClientIP)
	if pError != nil {
		log.Println("PrepareAccessTokens()::GetAccountId(E): " + pError.Error())
	}

	pError = IsExistsHistory(c.pDatabase, postgres.AuthorityEntra, c.pE.pId, c.pClientIP)
	if pError != nil {
		log.Println("PrepareAccessTokens()::GetAccountId(E): " + pError.Error())
	}

}

// PrepareSessionAccessToken セッションキーにアクセストークンを紐付け
func PrepareSessionAccessToken(pDatabase *sql.DB, eAuthority int, pSessionKey string, pAccessToken *oauth2.Token) error {
	// アクセストークンの存在を確認
	var pSToken oauth2.Token
	nRows, pError := GetSessionTokenByKey(pDatabase, eAuthority, pSessionKey, &pSToken)
	if pError != nil {
		fmt.Printf("PrepareAccessToken(%s)::GetTokenByKey() AuthorityId = %s, SessionKey = %s\n", postgres.AuthorityName[eAuthority], postgres.Authoritys[eAuthority], pSessionKey)
		return pError
	}
	if nRows == 0 {
		// 準正常：セッションキーに紐付くアクセストークンは存在しない。
		return nil
	}

	// アクセストークンを保存
	*pAccessToken = pSToken

	return nil
}

// PrepareDelegateAccessToken セッションキーにアクセストークンを紐付け
func PrepareDelegateAccessToken(pDatabase *sql.DB, pUniqueId string, eAuthority int, pAccessToken *oauth2.Token) error {
	// アクセストークンの存在を確認
	var pAToken oauth2.Token
	nRows, pError := GetDelegateTokenById(pDatabase, eAuthority, pUniqueId, &pAToken)
	if pError != nil {
		fmt.Printf("PrepareDelegateAccessToken(%s)::GetDelegateTokenById() AuthorityId = %s, UniqueId = %s\n", postgres.AuthorityName[eAuthority], postgres.Authoritys[eAuthority], pUniqueId)
		return pError
	}
	if nRows == 0 {
		// 準正常：セッションキーに紐付くアクセストークンは存在しない。
		return nil
	}

	// アクセストークンを保存
	*pAccessToken = pAToken

	return nil
}

// GetAccountId
func GetAccountId(pDatabase *sql.DB, pUniqueId string, eAuthority int, pAccountId *string) error {
	//
	pId := ""
	pError := postgres.GetAccountId(pDatabase, pUniqueId, eAuthority, &pId)
	if pError != nil {
		fmt.Printf("GetAccountId(%s)::GetAccountId() AuthorityId = %s\n", postgres.AuthorityName[eAuthority], postgres.Authoritys[eAuthority])
		return pError
	}
	*pAccountId = pId

	return nil
}

// IsExistsHistory
func IsExistsHistory(pDatabase *sql.DB, eAuthority int, pAccountId string, pClientIP string) error {
	// 前回通信したグローバルIPアドレスでなければ、異常と判定（判定条件が荒い）
	nRows, pError := postgres.IsExistsHistory(pDatabase, eAuthority, pAccountId, pClientIP)
	if pError != nil {
		return pError
	}
	if nRows == 0 {
		// 初回アクセス
	}

	return nil
}

// HandlerGooglePicture Googleプロファイルアイコン
func HandlerGooglePicture(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	log.Println("HandlerGooglePicture()")
	var pUserInfo googleapi.UserInfo
	pError := googleapi.GetUserInfoGoogle(r, &c.pG.pAToken, &pUserInfo)
	if pError != nil {
		log.Println("ERROR: HandlerGooglePicture().GetUserInfoGoogle() " + pError.Error())
		http.Redirect(w, r, "/error", http.StatusTemporaryRedirect)
		return
	}
	pPictureURL := pUserInfo.Picture
	pBytes, pContentType, pError := GetIconStream(r.Context(), pPictureURL, &c.pG.pAToken)
	if pError != nil {
		log.Println("ERROR: HandlerGooglePicture() " + pError.Error())
		http.Redirect(w, r, "/error", http.StatusTemporaryRedirect)
		return
	}
	w.Header().Set("Content-Type", pContentType)
	w.Header().Set("Cache-Control", "public, max-age=3600")
	_, pError = w.Write(pBytes)
	if pError != nil {
		log.Printf("%s", pError.Error())
		return
	}
}

// HandlerEntraPicture Entraプロファイルアイコン
func HandlerEntraPicture(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	log.Println("HandlerEntraPicture()")
	var pUserInfo entraapi.UserInfo
	pError := entraapi.GetUserInfoEntra(r, &c.pE.pAToken, &pUserInfo)
	if pError != nil {
		log.Println("ERROR: HandlerEntraPicture().GetUserInfoEntra() " + pError.Error())
		http.Redirect(w, r, "/error", http.StatusTemporaryRedirect)
		return
	}
	pPictureURL := fmt.Sprintf("https://graph.microsoft.com/v1.0/users/%s/photo/$value", pUserInfo.ID)
	pBytes, pContentType, pError := GetIconStream(r.Context(), pPictureURL, &c.pE.pAToken)
	if pError != nil {
		log.Println("ERROR: HandlerEntraPicture() " + pError.Error())
		http.Redirect(w, r, "/error", http.StatusTemporaryRedirect)
		return
	}
	w.Header().Set("Content-Type", pContentType)
	w.Header().Set("Cache-Control", "public, max-age=3600")
	_, pError = w.Write(pBytes)
	if pError != nil {
		log.Printf("%s", pError.Error())
		return
	}
}

// HandlerGoogleSubscribe
func HandlerGoogleSubscribe(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	// サブスクリプションを記録
	pResourceId := "me/events"
	pClientState := generateState()
	pAccountId := c.pG.pId

	pExpireTimeStamp := time.Now().UTC().Add(time.Minute * 45)
	log.Println("ExpireTimeStamp = " + pExpireTimeStamp.String())
	log.Printf("AccountId=[%s]\n", pAccountId)
	pSubscriptionId := ""
	googleapi.Subscribe(r, &c.pG.pAToken, pClientState, pExpireTimeStamp, &pSubscriptionId, &pResourceId)
	log.Println("SubscriptionId: " + pSubscriptionId)
	log.Println("UniqueId: " + c.pUniqueId)

	pError := postgres.CreateSubscribe(c.pDatabase, c.pUniqueId, postgres.AuthorityGoogle, pAccountId, pResourceId, pClientState, pExpireTimeStamp, pSubscriptionId)
	if pError != nil {
		log.Println("HandlerGoogleSubscribe(): " + pError.Error())
	}

	//　トップページにリダイレクト
	pURL := "/"
	http.Redirect(w, r, pURL, http.StatusTemporaryRedirect)
}

// HandlerEntraSubscribe サブスクリプションを登録
func HandlerEntraSubscribe(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	// サブスクリプションを記録
	pResourceId := "me/events"
	pClientState := generateState()
	pAccountId := c.pE.pId

	pExpireTimeStamp := time.Now().UTC().Add(time.Minute * 45)
	log.Println("ExpireTimeStamp = " + pExpireTimeStamp.String())
	log.Printf("AccountId=[%s]\n", pAccountId)
	pSubscriptionId := ""
	entraapi.Subscribe(r, &c.pE.pAToken, pResourceId, pClientState, pExpireTimeStamp, &pSubscriptionId)
	log.Println("SubscriptionId: " + pSubscriptionId)
	log.Println("UniqueId: " + c.pUniqueId)

	pError := postgres.CreateSubscribe(c.pDatabase, c.pUniqueId, postgres.AuthorityEntra, pAccountId, pResourceId, pClientState, pExpireTimeStamp, pSubscriptionId)
	if pError != nil {
		log.Println("HandlerEntraSubscribe(): " + pError.Error())
	}

	//　トップページにリダイレクト
	pURL := "/"
	http.Redirect(w, r, pURL, http.StatusTemporaryRedirect)
}

// HandlerCheckSubscribe サブスクリプションの有効状態を検査
func HandlerCheckSubscribe(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	//HandlerDuplicateGoogleCalendar(w, r, c)
	//HandlerDuplicateEntraCalendar(w, r, c)
	//　トップページにリダイレクト
	pURL := "/"
	http.Redirect(w, r, pURL, http.StatusTemporaryRedirect)
}

// HandlerUnsubscribe
func HandlerUnsubscribe(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	HandlerGoogleUnsubscribe(w, r, c)
	HandlerEntraUnsubscribe(w, r, c)
}

// HandlerGoogleUnsubscribe サブスクリプションを解除
func HandlerGoogleUnsubscribe(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	// サブスクリプションを登録解除
	pAccountId := c.pG.pId
	pResults, pError := postgres.ListupSubscribe(c.pDatabase, postgres.AuthorityGoogle, pAccountId)
	if pError != nil {
		return
	} else {
		//
		for _, pSubscribe := range pResults {
			pClientState := generateState()
			googleapi.Unsubscribe(r, &c.pG.pAToken, pSubscribe.SubscribeId, pSubscribe.ResourceId, pClientState)
			log.Printf("Unsubscribe:: SubscriptionId: %s, ResourceId: %s\n", pSubscribe.SubscribeId, pSubscribe.ResourceId)
			postgres.DeleteSubscribe(c.pDatabase, postgres.AuthorityGoogle, pAccountId, pSubscribe.SubscribeId)
		}
	}

	//　トップページにリダイレクト
	pURL := "/"
	http.Redirect(w, r, pURL, http.StatusTemporaryRedirect)
}

// HandlerGoogleCheckSubscribe
func HandlerGoogleCheckSubscribe(w http.ResponseWriter, r *http.Request, c *FunctionContext) {

}

// HandlerEntraCheckSubscribe サブスクリプションの有効状態を検査し、無効なものを削除する。
func HandlerEntraCheckSubscribe(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	pAccountId := c.pE.pId
	pResults, pError := postgres.ListupSubscribe(c.pDatabase, postgres.AuthorityEntra, pAccountId)
	if pError != nil {
		return
	} else {
		//
		for _, pSubscribe := range pResults {
			pError := entraapi.GetSubscribe(r, &c.pE.pAToken, pSubscribe.SubscribeId)
			if pError != nil {
				// サブスクリプションを照会した結果エラーの場合は、内部のサブスクリプション情報を削除する。
				postgres.DeleteSubscribe(c.pDatabase, postgres.AuthorityEntra, pAccountId, pSubscribe.SubscribeId)
			} else {
				log.Printf("Hit: HandlerGoogleCheckSubscribe().SubscriptionId: %s, ResourceId: %s\n", pSubscribe.SubscribeId, pSubscribe.ResourceId)
			}
		}
	}
}

// HandlerEntraUnsubscribe サブスクリプションを解除
func HandlerEntraUnsubscribe(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	pAccountId := c.pE.pId
	pResults, pError := postgres.ListupSubscribe(c.pDatabase, postgres.AuthorityEntra, pAccountId)
	if pError != nil {
		return
	} else {
		//
		for _, pSubscribe := range pResults {
			entraapi.Unsubscribe(r, &c.pE.pAToken, pSubscribe.SubscribeId)
			log.Printf("Unsubscribe:: SubscriptionId: %s, ResourceId: %s\n", pSubscribe.SubscribeId, pSubscribe.ResourceId)
			postgres.DeleteSubscribe(c.pDatabase, postgres.AuthorityEntra, pAccountId, pSubscribe.SubscribeId)
		}
	}

	//　トップページにリダイレクト
	pURL := "/"
	http.Redirect(w, r, pURL, http.StatusTemporaryRedirect)
}

// HandlerGoogleListupSubscribe
func HandlerGoogleListupSubscribe(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	// ページに出力する情報を収集
	var pSubscribes googleapi.SubscriptionListResponse
	pError := PrepareListupSubscribeGoogle(r, c, &pSubscribes)
	if pError != nil {
		log.Println(pError.Error())
		pURL := "/"
		http.Redirect(w, r, pURL, http.StatusTemporaryRedirect)
	} else {
		// ページを出力
		WriteHtmlHeaders(w, pApplicationName, http.StatusOK)

		fmt.Fprintln(w, "<body>")
		fmt.Fprintln(w, "<h2>サブスクリプション一覧</h2>")
		fmt.Fprintln(w, "<p>"+time.Now().Format(time.RFC1123Z)+"<br /></p>")

		fmt.Fprintln(w, "<hr />")
		fmt.Fprintf(w, `<div class="header">`)
		fmt.Fprintln(w, "<p>")
		fmt.Fprintln(w, "<a href='/'>戻る</a>")
		fmt.Fprintln(w, "</p>")
		fmt.Fprintf(w, `</div>`)

		fmt.Fprintln(w, "<p>")
		for _, pSubscribe := range pSubscribes.Value {
			fmt.Fprintf(w, "SubscribeId: %s, ResourceId: %s<br />", pSubscribe.ID, pSubscribe.Resource)
		}

		fmt.Fprintln(w, "</p>")

		fmt.Fprintln(w, "<hr />")
		fmt.Fprintln(w, `<footer>`)
		fmt.Fprintln(w, "<p>Copyright 2025 Arteria Studio, All right reserved. </p>")
		fmt.Fprintln(w, "</footer>")

		fmt.Fprintln(w, "</body>")
	}
}

// GetIconStream アイコンファイルをダウンロード
func GetIconStream(ctx context.Context, pURL string, pAToken *oauth2.Token) ([]byte, string, error) {
	// HTTPリクエスト
	req, err := http.NewRequestWithContext(ctx, "GET", pURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("リクエストの作成に失敗: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+pAToken.AccessToken)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("GetIconStream(): Graph APIの呼び出しに失敗: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// 写真がない場合はエラーではなくnilを返しても良い (ここでは簡略化のためエラー)
		return nil, "", fmt.Errorf("アイコンが見つかりません (HTTP 404)。ユーザーが写真を設定していない可能性があります。")
	}
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			// アクセストークンが無効
		}
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("GetIconStream(): Graph APIの呼び出しエラー: ステータスコード %d, レスポンス: %s", resp.StatusCode, string(bodyBytes))
	}

	// 画像データを読み取り
	photoBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("画像データの読み取りに失敗: %w", err)
	}

	// Content-Type をレスポンスヘッダーから取得
	contentType := resp.Header.Get("Content-Type")

	return photoBytes, contentType, nil
}

// PrepareSessionKey セッションキーを獲得または生成
func PrepareSessionKey(r *http.Request, c *FunctionContext) error {
	c.pSessionKey = PickupCookie(r, "XSRS-TOKEN")
	if c.pSessionKey == "" {
		c.pSessionKey = generateState()
	}
	pExpireStamp := time.Now().Add(time.Second * 86400)
	pError := postgres.CreateSessionKey(c.pDatabase, c.pSessionKey, pExpireStamp)
	if pError != nil {
		return pError
	}
	return nil
}

// HandlerSignin サインイン
func HandlerSignin(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	pMailad := r.FormValue("email")
	pPasswd := r.FormValue("password")
	pKey := pMailad + "#" + pPasswd
	pVal := GenerateSHA256Hash(pKey)

	nRows, pError := postgres.CountOfAccount(c.pDatabase, pMailad, pVal, &c.pUniqueId, &c.pDisplayName)
	if pError != nil {
		log.Println("HandlerSignin()::LookupAccount(): " + pError.Error())
	}
	if nRows == 0 {
		// 該当するアカウントは登録されていない。
		log.Println("HandlerSignin(): Account not found. " + pMailad)
	} else {
		// セッション情報を登録
		pError := postgres.UpdateSessionKey(c.pDatabase, c.pSessionKey, c.pUniqueId)
		if pError != nil {
			log.Println("HandlerSignin()::UpdateSessionKey(): " + pError.Error())
		} else {
			// ログインを承認。アクセスを許可
			log.Printf("SUCCESS: HandlerSignin() e-Mail: %s, DisplayName: %s\n", pMailad, c.pDisplayName)
		}
	}
	//　トップページにリダイレクト
	pURL := "/"
	http.Redirect(w, r, pURL, http.StatusTemporaryRedirect)
}

// HandlerSignup サインアップ
func HandlerSignup(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	pMailad := r.FormValue("email")
	pPasswdRegist := r.FormValue("registpassword")
	pPasswdVerify := r.FormValue("verifypassword")
	pKey := pMailad + "#" + pPasswdRegist
	pVal := GenerateSHA256Hash(pKey)
	if pPasswdRegist == pPasswdVerify {
		pError := postgres.RegistAccount(c.pDatabase, pMailad, pVal, pMailad)
		if pError != nil {
			log.Println("HandlerSignup(): " + pError.Error())
		}
	}

	//　トップページにリダイレクト
	pURL := "/"
	http.Redirect(w, r, pURL, http.StatusTemporaryRedirect)
}

// HandlerPageSignup サインアップページ
func HandlerPageSignup(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	WriteHtmlHeaders(w, pApplicationName, http.StatusOK)

	// CSRF対策の仕掛け（未実装）
	pCSRF_Token := generateState()

	fmt.Fprintln(w, "<body>")
	fmt.Fprintln(w, "<h2>サインアップ（arteria-s.net/page/signup）</h2>")
	fmt.Fprintln(w, "<hr />")
	fmt.Fprintln(w, `<div class="login-container">`)
	fmt.Fprintln(w, `<form action="/auth/signup" method="post">`)
	fmt.Fprintln(w, `<div class="input-group">`)
	fmt.Fprintln(w, `<label for="email">メールアドレス:</label>`)
	fmt.Fprintln(w, `<input type="email" id="email" name="email" required placeholder="your.email@example.com">`)
	fmt.Fprintln(w, `</div>`)
	fmt.Fprintln(w, `<div class="input-group">`)
	fmt.Fprintln(w, `<label for="password">パスワード:</label>`)
	fmt.Fprintln(w, `<input type="password" id="registpassword" name="registpassword" required placeholder="********">`)
	fmt.Fprintln(w, `<label for="verifypassword">確認用パスワード:</label>`)
	fmt.Fprintln(w, `<input type="password" id="verifypassword" name="verifypassword" required placeholder="********">`)
	fmt.Fprintf(w, `<input type="hidden" name="csrf_token" value="%s">`, pCSRF_Token)
	fmt.Fprintln(w, `</div>`)
	fmt.Fprintln(w, `<button type="submit">サインアップ</button>`)
	fmt.Fprintln(w, `</form>`)
	fmt.Fprintln(w, `</div>`)
	fmt.Fprintln(w, "<hr />")
	fmt.Fprintln(w, `<p><a href="/">トップページに戻る。</p>`)
	fmt.Fprintln(w, "</body>")
}
