package signin

import (
	"log"
	"net/http"

	"arteria-s.net/entraapi"
	"arteria-s.net/googleapi"
	"arteria-s.net/postgres"
	"golang.org/x/oauth2"
)

// HandlerGoogleSignin 対話ログイン（Google）
func HandlerGoogleSignin(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	// アカウントに対するアプリケーションからのアクセスを許諾するパーミッションを要求
	var pToken oauth2.Token
	nRows, pError := GetSessionTokenByKey(c.pDatabase, postgres.AuthorityGoogle, c.pSessionKey, &pToken)
	if (pError != nil) || (nRows > 0) {
		log.Println("パーミッションを獲得済み")
		// パーミッションを獲得済み
		pURL := "/"
		http.Redirect(w, r, pURL, http.StatusTemporaryRedirect)
	} else {
		// pStateはCSRF攻撃を防ぐために使用されるランダムな文字列
		pChallenge := generateState()
		pError := postgres.CreateChallenge(c.pDatabase, c.pSessionKey, pChallenge)
		if pError != nil {
			// チャレンジ値の記録に失敗（安全性を担保できないのでエラーへリダイレクト）
			pURL := "/auth/error"
			http.Redirect(w, r, pURL, http.StatusTemporaryRedirect)
		} else {
			// 対話型アクセストークンの獲得シーケンスを開始
			googleapi.HandleGoogleLogin(w, r, pChallenge)
		}
	}
}

// HandlerGoogleAuthResponse 認証後処理
func HandlerGoogleAuthResponse(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	pURL := "/"

	// stateの検証（stateをわざわざクエリーパラメータに配置する設計が秀逸だと受け止める。）
	pQueryParam := r.URL.Query()
	pChallenge := pQueryParam.Get("state")
	nRows, _ := postgres.IsExistsChallenge(c.pDatabase, c.pSessionKey, pChallenge)
	if nRows <= 0 {
		// 記録にないチャレンジなのでリクエストを捨てる。
		log.Println("WARNING: CSRFと思われるリクエストを拒否 " + pChallenge)
		pURL = "/auth/error"
	} else {
		// チャレンジ値の記録を削除
		_, pError := postgres.DeleteChallenge(c.pDatabase, c.pSessionKey, pChallenge)
		if pError != nil {
			// 前段のクエリーが成功しておきながら削除が失敗する可能性は並行処理環境だと想定できるがエラーは無視する。
			log.Println("WARNING: 時間経過で前段のクエリーと矛盾する結果が発生。エラーは無視（DeleteChallenge）")
		}

		// 認証コードからアクセストークンを交換
		pError = googleapi.HandleGoogleCallback(w, r, &c.pG.pSToken)
		if pError != nil {
			// エラー
			pURL = "/auth/error"
		} else {
			// アクセストークンを記録
			pError := SetSessionToken(c.pDatabase, postgres.AuthorityGoogle, c.pSessionKey, &c.pG.pSToken)
			if pError != nil {
				pURL = "/auth/error"
			}

			// 利用者識別子を獲得
			var pUserInfo googleapi.UserInfo
			googleapi.GetUserInfoGoogle(r, &c.pG.pSToken, &pUserInfo)

			// アカウント台帳からユニーク識別子を入力
			pAccountId := pUserInfo.Email
			pError = postgres.GetUniqueId(c.pDatabase, postgres.AuthorityGoogle, pAccountId, &c.pUniqueId)
			if pError != nil {
				log.Printf("FAILED: GetUniqueId(): %s\n", pError.Error())
			}

			// セッションキーを記録
			pError = postgres.UpdateSessionKey(c.pDatabase, c.pSessionKey, c.pUniqueId)
			if pError != nil {
				// エラー
				pURL = "/auth/error"
				log.Printf("FAILED: %s\n", pError.Error())
				log.Printf("FAILED: HandlerGoogleAuthResponse(): UniqueId=[%s], SessionKey=[%s]\n", c.pUniqueId, c.pSessionKey)
			}

			// ログイン成功履歴を記録
			pError = postgres.AddLoginHistory(c.pDatabase, c.pUniqueId, postgres.AuthorityGoogle, pAccountId, c.pClientIP)
			if pError != nil {
				// エラー
				pURL = "/auth/error"
				log.Printf("FAILED: HandlerGoogleAuthResponse()::AddLoginHistory() UniqueId=[%s], SessionKey=[%s]\n", c.pUniqueId, c.pSessionKey)
				log.Printf("FAILED: %s\n", pError.Error())
			}

		}
	}

	http.Redirect(w, r, pURL, http.StatusTemporaryRedirect)
}

// HandlerEntraSignin 対話ログイン（Entra）
func HandlerEntraSignin(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	// アカウントに対するアプリケーションからのアクセスを許諾するパーミッションを要求
	var pToken oauth2.Token
	nRows, pError := GetSessionTokenByKey(c.pDatabase, postgres.AuthorityEntra, c.pSessionKey, &pToken)
	if (pError != nil) || (nRows > 0) {
		// エラーが発生したときはサインイン処理を中断してデフォルトページに遷移する。
		pURL := "/"
		http.Redirect(w, r, pURL, http.StatusTemporaryRedirect)
	} else {
		// pStateはCSRF攻撃を防ぐために使用されるランダムな文字列
		pChallenge := generateState()
		pError := postgres.CreateChallenge(c.pDatabase, c.pSessionKey, pChallenge)
		if pError != nil {
			// チャレンジ値の記録に失敗（安全性を担保できないのでエラーへリダイレクト）
			pURL := "/auth/error"
			http.Redirect(w, r, pURL, http.StatusTemporaryRedirect)
		} else {
			// パーミッションを獲得するためのシーケンスを開始
			entraapi.HandleEntraLogin(w, r, pChallenge)
		}
	}
}

// HandlerEntraAuthResponse 認証後処理
func HandlerEntraAuthResponse(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	pURL := "/"

	// stateの検証（stateをわざわざクエリーパラメータに配置する設計が秀逸だと受け止める。）
	pQueryParam := r.URL.Query()
	pChallenge := pQueryParam.Get("state")
	nRows, pError := postgres.IsExistsChallenge(c.pDatabase, c.pSessionKey, pChallenge)
	if pError != nil {
		// エラー
		pURL = "/auth/error"
	} else {
		if nRows <= 0 {
			// 記録にないチャレンジなのでリクエストを捨てる。
			log.Println("WARNING: CSRFと思われるリクエストを拒否 " + pChallenge)
			pURL = "/auth/error"
		} else {
			// チャレンジ値の記録を削除
			_, pError := postgres.DeleteChallenge(c.pDatabase, c.pSessionKey, pChallenge)
			if pError != nil {
				// 前段のクエリーが成功しておきながら削除が失敗する可能性は並行処理環境だと想定できるがエラーは無視する。
				log.Println("WARNING: 時間経過で前段のクエリーと矛盾する結果が発生。エラーは無視（DeleteChallenge）")
			}

			// 認証コードからアクセストークンを交換
			pError = entraapi.HandleEntraCallback(w, r, &c.pE.pAToken)
			if pError != nil {
				// エラー
				pURL = "/auth/error"
			} else {
				// アクセストークンを記録
				pError := SetSessionToken(c.pDatabase, postgres.AuthorityEntra, c.pSessionKey, &c.pE.pAToken)
				if pError != nil {
					pURL = "/auth/error"
				}

				// 利用者識別子を獲得
				var pUserInfo entraapi.UserInfo
				entraapi.GetUserInfoEntra(r, &c.pE.pAToken, &pUserInfo)

				// アカウント台帳からユニーク識別子を入力
				pAccountId := pUserInfo.Email
				pError = postgres.GetUniqueId(c.pDatabase, postgres.AuthorityEntra, pAccountId, &c.pUniqueId)
				if pError != nil {
					log.Printf("FAILED: HandlerEntraAuthResponse()::GetUniqueId(): %s\n", pError.Error())
				}

				// セッションキーを記録
				pError = postgres.UpdateSessionKey(c.pDatabase, c.pSessionKey, c.pUniqueId)
				if pError != nil {
					// エラー
					pURL = "/auth/error"
					log.Printf("FAILED: %s\n", pError.Error())
					log.Printf("FAILED: HandlerEntraAuthResponse()::UpdateSessionKey() UniqueId=[%s], SessionKey=[%s]\n", c.pUniqueId, c.pSessionKey)
				}

				// ログイン成功履歴を記録
				pError = postgres.AddLoginHistory(c.pDatabase, c.pUniqueId, postgres.AuthorityEntra, pAccountId, c.pClientIP)
				if pError != nil {
					// エラー
					pURL = "/auth/error"
					log.Printf("FAILED: HandlerEntraAuthResponse()::AddLoginHistory() UniqueId=[%s], SessionKey=[%s]\n", c.pUniqueId, c.pSessionKey)
					log.Printf("FAILED: %s\n", pError.Error())
				}
			}
		}
	}

	http.Redirect(w, r, pURL, http.StatusTemporaryRedirect)
}
