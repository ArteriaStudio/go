package entraapi

import (
	"encoding/json"
	"io"
	"net/http"

	"golang.org/x/oauth2"
)

// GetUserInfoEntra ユーザー情報
func GetUserInfoEntra(r *http.Request, pAccessToken *oauth2.Token, pUserInfo *UserInfo) error {
	// Access Tokenを使用してユーザー情報を取得
	pUserInfoEndpoint := "https://graph.microsoft.com/v1.0/me"
	pClient := pEntraOauthConfig.Client(r.Context(), pAccessToken)
	if pClient != nil {
		resp, pError := pClient.Get(pUserInfoEndpoint)
		if pError != nil {
			return pError
		}
		defer resp.Body.Close()

		// レスポンスからユーザー情報をパース
		pBody, pError := io.ReadAll(resp.Body)
		if pError != nil {
			return pError
		}

		pError = json.Unmarshal(pBody, pUserInfo)
		if pError != nil {
			return pError
		}
	}

	return nil
}
