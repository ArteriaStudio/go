package entraapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

// Graph APIに送信するサブスクリプション要求の構造体
type SubscriptionRequest struct {
	ChangeType               string `json:"changeType"`
	NotificationURL          string `json:"notificationUrl"`
	LifecycleNotificationUrl string `json:"lifecycleNotificationUrl"`
	Resource                 string `json:"resource"`
	ExpirationDateTime       string `json:"expirationDateTime"`
	ClientState              string `json:"clientState"` // 任意の文字列。通知の検証に使用
}

// Graph APIに送信するサブスクリプション応答の構造体
type SubscriptionResponse struct {
	Id                       string `json:"id"`
	Resource                 string `json:"resource"`
	ApplicationId            string `json:"ApplicationId"`
	ChangeType               string `json:"changeType"`
	ClientState              string `json:"clientState"` // 任意の文字列。通知の検証に使用
	NotificationURL          string `json:"notificationUrl"`
	LifecycleNotificationUrl string `json:"lifecycleNotificationUrl"`
	ExpirationDateTime       string `json:"expirationDateTime"`
}

//【サブスクリプションのエンドポイント】
// https://learn.microsoft.com/ja-jp/microsoft-365/enterprise/additional-office365-ip-addresses-and-urls?view=o365-worldwide

// Subscribe イベントをフック
func Subscribe(r *http.Request, pAccessToken *oauth2.Token, pResourceId string, pClientState string, pExpire time.Time, pSubscriptionId *string) {
	// 有効期限は、現在時刻から最大で約70.5時間後 (4230分)
	pExpirationDateTime := pExpire.Format(time.RFC3339)
	log.Println("pExpirationDateTime = " + pExpirationDateTime)

	//
	pSubscription := SubscriptionRequest{
		ChangeType:               "created,updated,deleted",                                 // 変更を監視する操作
		NotificationURL:          "https://api.arteria-s.net:8443/listener/entra",           // 公開アクセス可能なWebhookエンドポイント
		LifecycleNotificationUrl: "https://api.arteria-s.net:8443/listener/entra/lifecycle", // ライフサイクル
		Resource:                 pResourceId,                                               // 監視対象のリソース
		ExpirationDateTime:       pExpirationDateTime,
		ClientState:              pClientState, // 通知が本物か検証するための秘密の文字列
	}

	// JSON形式に構造体を変換
	pRequestBody, err := json.Marshal(pSubscription)
	if err != nil {
		fmt.Printf("Error marshalling JSON: %v\n", err)
		return
	}

	// WebHook エンドポイント
	pEndPoint := "https://graph.microsoft.com/v1.0/subscriptions"
	pRequest, pError := http.NewRequest("POST", pEndPoint, bytes.NewBuffer(pRequestBody))
	if pError != nil {
		log.Printf("http.NewRequest(): %s", pError.Error())
		return
	}
	pRequest.Header.Set("Authorization", "Bearer "+pAccessToken.AccessToken)
	pRequest.Header.Set("Content-Type", "application/json")

	// HTTPリクエストを送信
	pClient := &http.Client{Timeout: 10 * time.Second}
	pResponse, pError := pClient.Do(pRequest)
	if pError != nil {
		log.Printf("http.Client().Do(): %s", pError.Error())
		return
	}
	defer pResponse.Body.Close()

	if pResponse.StatusCode != http.StatusCreated {
		// エラーレスポンスのボディを読み取って詳細を出力できます
		bodyBytes, _ := io.ReadAll(pResponse.Body)
		log.Printf("Subscribe() ==> %s", string(bodyBytes))
		return
	}

	var pSubscriptionResponse SubscriptionResponse
	pError = json.NewDecoder(pResponse.Body).Decode(&pSubscriptionResponse)
	if pError != nil {
		log.Printf("JSON デコードエラー: %v", pError)
		return
	}

	//
	*pSubscriptionId = pSubscriptionResponse.Id
}

// Unsubscribe イベントフックを解除
func Unsubscribe(r *http.Request, pAccessToken *oauth2.Token, pSubscribeId string) {
	// WebHook エンドポイント
	pEndPoint := "https://graph.microsoft.com/v1.0/subscriptions/" + pSubscribeId
	pRequest, pError := http.NewRequest("DELETE", pEndPoint, nil)
	if pError != nil {
		log.Printf("http.NewRequest(): %s", pError.Error())
		return
	}
	pRequest.Header.Set("Authorization", "Bearer "+pAccessToken.AccessToken)
	pRequest.Header.Set("Content-Type", "application/json")

	// HTTPリクエストを送信
	pClient := &http.Client{Timeout: 10 * time.Second}
	pResponse, pError := pClient.Do(pRequest)
	if pError != nil {
		log.Printf("http.Client().Do(): %s", pError.Error())
		return
	}
	defer pResponse.Body.Close()

	if pResponse.StatusCode != http.StatusNoContent {
		// エラーレスポンスのボディを読み取って詳細を出力できます
		bodyBytes, _ := io.ReadAll(pResponse.Body)
		log.Printf("Unsubscribe() ==> %s", string(bodyBytes))
		return
	}
}

// SubscribeUpdate サブスクリプションの有効期限を更新
func SubscribeUpdate(r *http.Request, pAccessToken *oauth2.Token, pExpire time.Time, pSubscriptionId string) error {
	// 有効期限は、現在時刻から最大で約70.5時間後 (4230分)
	pExpirationDateTime := pExpire.Format(time.RFC3339)
	log.Println("pExpirationDateTime = " + pExpirationDateTime)

	pSubscription := SubscriptionRequest{
		ExpirationDateTime: pExpirationDateTime,
	}

	// JSON形式に構造体を変換
	pRequestBody, pError := json.Marshal(pSubscription)
	if pError != nil {
		fmt.Printf("Error marshalling JSON: %v\n", pError)
		return pError
	}

	// WebHook エンドポイント
	pEndPoint := "https://graph.microsoft.com/v1.0/subscriptions/" + pSubscriptionId
	pRequest, pError := http.NewRequest("PATCH", pEndPoint, bytes.NewBuffer(pRequestBody))
	if pError != nil {
		log.Printf("http.NewRequest(): %s", pError.Error())
		return pError
	}
	pRequest.Header.Set("Authorization", "Bearer "+pAccessToken.AccessToken)
	pRequest.Header.Set("Content-Type", "application/json")

	// HTTPリクエストを送信
	pClient := &http.Client{Timeout: 10 * time.Second}
	pResponse, pError := pClient.Do(pRequest)
	if pError != nil {
		log.Printf("http.Client().Do(): %s", pError.Error())
		return pError
	}
	defer pResponse.Body.Close()

	if pResponse.StatusCode != http.StatusOK {
		// エラーレスポンスのボディを読み取って詳細を出力できます
		bodyBytes, _ := io.ReadAll(pResponse.Body)
		log.Printf("SubscribeUpdate() ==> %s", string(bodyBytes))
		return fmt.Errorf("SubscribeUpdate(): StatusCode=%d", pResponse.StatusCode)
	}

	return nil
}
