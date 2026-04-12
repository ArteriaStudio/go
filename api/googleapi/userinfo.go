package googleapi

import (
	"encoding/json"
	"io"
	"net/http"

	"golang.org/x/oauth2"
)

// GetUserInfoGoogle ユーザー情報
func GetUserInfoGoogle(r *http.Request, pAccessToken *oauth2.Token, pUserInfo *UserInfo) error {
	// Access Tokenを使用してユーザー情報を取得
	userInfoEndpoint := "https://www.googleapis.com/oauth2/v2/userinfo"
	pClient := googleOauthConfig.Client(r.Context(), pAccessToken)
	if pClient != nil {
		resp, pError := pClient.Get(userInfoEndpoint)
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
