package signin

import (
	"log"
	"net/http"

	"arteria-s.net/entraapi"
	"arteria-s.net/googleapi"
	"arteria-s.net/postgres"
	"golang.org/x/oauth2"
)

// HandlerGoogleGrant サインイン処理
func HandlerGoogleGrant(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	// アカウントに対するアプリケーションからのアクセスを許諾するパーミッションを要求
	var pToken oauth2.Token
	nRows, pError := GetSessionTokenByKey(c.pDatabase, postgres.AuthorityGoogle, c.pUniqueId, &pToken)
	if (pError != nil) || (nRows > 0) {
		log.Println("パーミッションを獲得済み")
		// パーミッションを獲得済み
		pURL := "/"
		http.Redirect(w, r, pURL, http.StatusTemporaryRedirect)
	} else {
		// pStateはCSRF攻撃を防ぐために使用されるランダムな文字列
		pChallenge := generateState()
		pError := postgres.CreateChallenge(c.pDatabase, c.pUniqueId, pChallenge)
		if pError != nil {
			// チャレンジ値の記録に失敗（安全性を担保できないのでエラーへリダイレクト）
			pURL := "/auth/error"
			http.Redirect(w, r, pURL, http.StatusTemporaryRedirect)
		} else {
			// パーミッションを獲得するためのシーケンスを開始
			googleapi.HandleGoogleGrant(w, r, pChallenge)
		}
	}
}

// HandlerGoogleRevoke アクセストークン破棄処理（Google）
func HandlerGoogleRevoke(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	log.Println("HandlerGoogleRevoke(): UniqueId: " + c.pUniqueId)

	// セッションキーからアクセストークンを獲得
	var pToken oauth2.Token
	nRows, pError := GetDelegateTokenById(c.pDatabase, postgres.AuthorityGoogle, c.pUniqueId, &pToken)
	if (pError != nil) || (nRows == 0) {
		// アクセストークンがない（アカウントに対するアクセスに関するパーミッションを獲得していない）
		pURL := "/"
		http.Redirect(w, r, pURL, http.StatusTemporaryRedirect)
	} else {
		//　アクセストークンを破棄
		googleapi.HandleGoogleRevoke(w, r, pToken.AccessToken)

		// アクセストークンを削除
		pError := DeleteDelegateToken(c.pDatabase, postgres.AuthorityGoogle, c.pUniqueId)
		if pError != nil {
			//
			log.Println("ERROR: HandlerGoogleRevoke().DeleteDelegateToken(): " + pError.Error())
		} else {
			log.Println("SUCCESS: HandlerGoogleRevoke()")
		}

		// ログアウト処理が完了したことを確認（レスポンスボディは通常空または無視される）
		http.Redirect(w, r, "/auth/logout", http.StatusTemporaryRedirect)
	}
}

// HandlerGooglePermResponse 認証後処理
func HandlerGooglePermResponse(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	pURL := "/"

	// stateの検証（stateをわざわざクエリーパラメータに配置する設計が秀逸だと受け止める。）
	pQueryParam := r.URL.Query()
	pChallenge := pQueryParam.Get("state")
	nRows, _ := postgres.IsExistsChallenge(c.pDatabase, c.pUniqueId, pChallenge)
	if nRows <= 0 {
		// 記録にないチャレンジなのでリクエストを捨てる。
		log.Println("WARNING: CSRFと思われるリクエストを拒否 " + pChallenge)
		pURL = "/auth/error"
	} else {
		// チャレンジ値の記録を削除
		_, pError := postgres.DeleteChallenge(c.pDatabase, c.pUniqueId, pChallenge)
		if pError != nil {
			// 前段のクエリーが成功しておきながら削除が失敗する可能性は並行処理環境だと想定できるがエラーは無視する。
			log.Println("WARNING: 時間経過で前段のクエリーと矛盾する結果が発生。エラーは無視（DeleteChallenge）")
		}

		// 認証コードからアクセストークンを交換
		pError = googleapi.HandleGoogleCallback(w, r, &c.pG.pAToken)
		if pError != nil {
			// エラー
			pURL = "/auth/error"
		} else {
			// アクセストークンを記録
			log.Printf("SetDelegateToken() R=[%s]\n", c.pG.pAToken.RefreshToken)
			pError := SetDelegateToken(c.pDatabase, postgres.AuthorityGoogle, c.pUniqueId, &c.pG.pAToken)
			if pError != nil {
				pURL = "/auth/error"
			}

			// 利用者識別子を獲得
			var pUserInfo googleapi.UserInfo
			pError = googleapi.GetUserInfoGoogle(r, &c.pG.pAToken, &pUserInfo)
			if pError != nil {
				// エラー
				pURL = "/auth/error"
				log.Printf("FAILED: %s\n", pError.Error())
				log.Printf("FAILED: HandlerGoogleAuthResponse().GetUserInfoGoogle(): UniqueId=[%s]\n", c.pUniqueId)
			}

			// アクセスリンクを更新
			pError = postgres.UpsertAccessLink(c.pDatabase, c.pUniqueId, postgres.AuthorityGoogle, pUserInfo.Email)
			if pError != nil {
				pURL = "/auth/error"
				log.Printf("FAILED: HandlerGooglePermResponse(): %s\n", pError.Error())
			}

			// ログイン成功履歴を記録
			pError = postgres.AddLoginHistory(c.pDatabase, c.pUniqueId, postgres.AuthorityGoogle, pUserInfo.Email, c.pClientIP)
			if pError != nil {
				// エラー
				pURL = "/auth/error"
				log.Printf("FAILED: HandlerGoogleAuthResponse()::AddLoginHistory() UniqueId=[%s]\n", c.pUniqueId)
				log.Printf("FAILED: %s\n", pError.Error())
			}

		}
	}

	http.Redirect(w, r, pURL, http.StatusTemporaryRedirect)
}

// HandlerEntraGrant アクセス許可申請
func HandlerEntraGrant(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	// アカウントに対するアプリケーションからのアクセスを許諾するパーミッションを要求
	var pToken oauth2.Token
	nRows, pError := GetDelegateTokenById(c.pDatabase, postgres.AuthorityEntra, c.pUniqueId, &pToken)
	if (pError != nil) || (nRows > 0) {
		// エラーが発生したときはサインイン処理を中断してデフォルトページに遷移する。
		pURL := "/"
		http.Redirect(w, r, pURL, http.StatusTemporaryRedirect)
	} else {
		// pStateはCSRF攻撃を防ぐために使用されるランダムな文字列
		pChallenge := generateState()
		pError := postgres.CreateChallenge(c.pDatabase, c.pUniqueId, pChallenge)
		if pError != nil {
			// チャレンジ値の記録に失敗（安全性を担保できないのでエラーへリダイレクト）
			pURL := "/auth/error"
			http.Redirect(w, r, pURL, http.StatusTemporaryRedirect)
		} else {
			// パーミッションを獲得するためのシーケンスを開始
			entraapi.HandleEntraGrant(w, r, pChallenge)
		}
	}
}

// HandlerEntraRevoke アクセストークン破棄処理（Entra）
func HandlerEntraRevoke(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	log.Println("HandlerEntraRevoke(): UniqueId: " + c.pUniqueId)

	// ユニークキーからアクセストークンを獲得
	var pToken oauth2.Token
	nRows, pError := GetDelegateTokenById(c.pDatabase, postgres.AuthorityEntra, c.pUniqueId, &pToken)
	if (pError != nil) || (nRows == 0) {
		// アクセストークンがない（アカウントに対するアクセスに関するパーミッションを獲得していない）
		pURL := "/"
		http.Redirect(w, r, pURL, http.StatusTemporaryRedirect)
	} else {
		//　アクセストークンを破棄
		var pResult bool = false
		entraapi.HandleEntraRevoke(r, &pToken, &pResult)

		// アクセストークンを削除
		pError := DeleteDelegateToken(c.pDatabase, postgres.AuthorityEntra, c.pUniqueId)
		if pError != nil {
			//
			log.Println("ERROR: HandlerEntraRevoke().DeleteDelegateToken(): " + pError.Error())
		} else {
			log.Println("SUCCESS: HandlerEntraRevoke()")
		}

		//　トップページにリダイレクト
		pURL := "/"
		http.Redirect(w, r, pURL, http.StatusTemporaryRedirect)
	}
}

// HandlerEntraPermResponse 認証後処理
func HandlerEntraPermResponse(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	log.Println("HandlerEntraPermResponse")
	pURL := "/"

	// stateの検証（stateをわざわざクエリーパラメータに配置する設計が秀逸だと受け止める。）
	pQueryParam := r.URL.Query()
	pChallenge := pQueryParam.Get("state")
	nRows, pError := postgres.IsExistsChallenge(c.pDatabase, c.pUniqueId, pChallenge)
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
			_, pError := postgres.DeleteChallenge(c.pDatabase, c.pUniqueId, pChallenge)
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
				pError := SetDelegateToken(c.pDatabase, postgres.AuthorityEntra, c.pUniqueId, &c.pE.pAToken)
				if pError != nil {
					pURL = "/auth/error"
				}

				// 利用者識別子を獲得
				var pUserInfo entraapi.UserInfo
				pError = entraapi.GetUserInfoEntra(r, &c.pE.pAToken, &pUserInfo)
				if pError != nil {
					// エラー
					pURL = "/auth/error"
					log.Printf("FAILED: HandlerEntraPermResponse()::GetUserInfoEntra() UniqueId=[%s]\n", c.pUniqueId)
					log.Printf("FAILED: %s\n", pError.Error())
				}

				// アクセスリンクを更新
				pError = postgres.UpsertAccessLink(c.pDatabase, c.pUniqueId, postgres.AuthorityEntra, pUserInfo.Email)
				if pError != nil {
					pURL = "/auth/error"
					log.Printf("FAILED: HandlerEntraPermResponse(): %s\n", pError.Error())
				}

				// ログイン成功履歴を記録
				pError = postgres.AddLoginHistory(c.pDatabase, c.pUniqueId, postgres.AuthorityEntra, pUserInfo.Email, c.pClientIP)
				if pError != nil {
					// エラー
					pURL = "/auth/error"
					log.Printf("FAILED: HandlerEntraPermResponse()::AddLoginHistory() UniqueId=[%s]\n", c.pUniqueId)
					log.Printf("FAILED: %s\n", pError.Error())
				}
			}
		}
	}

	http.Redirect(w, r, pURL, http.StatusTemporaryRedirect)
}
