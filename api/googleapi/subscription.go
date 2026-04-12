package googleapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

// SubscribeRequestParam
type SubscribeRequestParam struct {
	Ttl string `json:"ttl"`
}

// SubscriptionRequest Graph APIに送信するサブスクリプション要求の構造体
type SubscriptionRequest struct {
	Id      string                `json:"id"`
	Token   string                `json:"token"` // クライアントシークレット
	Type    string                `json:"type"`
	Address string                `json:"address"`
	Params  SubscribeRequestParam `json:"params"`
}

// SubscriptionResponse Graph APIに送信するサブスクリプション応答の構造体
type SubscriptionResponse struct {
	Id          string `json:"id"`
	ResourceId  string `json:"resourceId"`
	ResourceUri string `json:"resourceUri"`
	Token       string `json:"token"`
	Expiration  string `json:"expiration"`
}

// ChannelRequest チャネル識別情報
type ChannelRequest struct {
	Id         string `json:"id"`
	ResourceId string `json:"resourceId"`
	Token      string `json:"token"`
}

// SubscriptionListResponse は /subscriptions エンドポイントからの応答全体を表します。
type SubscriptionListResponse struct {
	OdataContext string         `json:"@odata.context"`
	Value        []Subscription `json:"value"`
}

// Subscription は個々のWebhookサブスクリプションを表します。
type Subscription struct {
	ID                          string    `json:"id"`
	Resource                    string    `json:"resource"`
	ApplicationID               string    `json:"applicationId"`
	ChangeType                  string    `json:"changeType"`
	ClientState                 string    `json:"clientState"`
	NotificationURL             string    `json:"notificationUrl"`
	NotificationQueryParameters string    `json:"notificationQueryParameters,omitempty"` // v2.0 endpoint
	LifecycleNotificationURL    string    `json:"lifecycleNotificationUrl,omitempty"`
	ExpirationDateTime          time.Time `json:"expirationDateTime"` // 有効期限
	CreatorID                   string    `json:"creatorId,omitempty"`
	LatestSupportedTlsVersion   string    `json:"latestSupportedTlsVersion,omitempty"`
	IncludeResourceData         bool      `json:"includeResourceData,omitempty"`
	TenantID                    string    `json:"tenantId,omitempty"`
}

// Subscribe イベントをフック
func Subscribe(r *http.Request, pAccessToken *oauth2.Token, pClientState string, pExpire time.Time, pSubscriptionId *string, pResourceId *string) {
	// 有効期限は、現在時刻から最大で約70.5時間後 (4230分)
	pExpirationDateTime := pExpire.Format(time.RFC3339)
	log.Println("pExpirationDateTime = " + pExpirationDateTime)

	pChannelId := uuid.New().String()

	//
	pSubscription := SubscriptionRequest{
		Id:      pChannelId,
		Token:   pClientState, // 通知が本物か検証するための秘密の文字列
		Type:    "webhook",
		Address: "https://api.arteria-s.net:8443/listener/google",
	}
	pSubscription.Params.Ttl = "1000"

	// JSON形式に構造体を変換
	pRequestBody, pError := json.Marshal(pSubscription)
	if pError != nil {
		fmt.Printf("Error marshalling JSON: %v\n", pError)
		return
	}

	// WebHook エンドポイント
	pEndPoint := "https://www.googleapis.com/calendar/v3/calendars/primary/events/watch"
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

	if pResponse.StatusCode != http.StatusOK {
		// エラーレスポンスのボディを読み取って詳細を出力できます
		bodyBytes, _ := io.ReadAll(pResponse.Body)
		log.Printf("GoogleAPI.Subscribe()[%d] ==> %s", pResponse.StatusCode, string(bodyBytes))
		return
	}

	var pSubscriptionResponse SubscriptionResponse
	pError = json.NewDecoder(pResponse.Body).Decode(&pSubscriptionResponse)
	if pError != nil {
		log.Printf("GoogleAPI.Subscribe() FAILED Decode JSON: %v", pError)
		return
	}

	//
	*pSubscriptionId = pSubscriptionResponse.Id
	*pResourceId = pSubscriptionResponse.ResourceId
}

// Unsubscribe イベントフックを解除
func Unsubscribe(r *http.Request, pAccessToken *oauth2.Token, pSubscriptionId string, pResourceId string, pClientState string) {

	pChannelRequest := ChannelRequest{
		Id:         pSubscriptionId,
		Token:      pClientState, // 通知が本物か検証するための秘密の文字列
		ResourceId: pResourceId,
	}
	// JSON形式に構造体を変換
	pRequestBody, pError := json.Marshal(pChannelRequest)
	if pError != nil {
		fmt.Printf("Error marshalling JSON: %v\n", pError)
		return
	}

	// WebHook エンドポイント
	pEndPoint := "https://www.googleapis.com/calendar/v3/channels/stop"
	pRequest, pError := http.NewRequest("DELETE", pEndPoint, bytes.NewBuffer(pRequestBody))
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
		log.Printf("GoogleAPI.Unsubscribe() ==> %s", string(bodyBytes))
		return
	}
}

// SubscribeUpdate サブスクリプションの有効期限を更新
func SubscribeUpdate(r *http.Request, pAccessToken *oauth2.Token, pExpire time.Time, pSubscriptionId string) error {
	return nil
}

// ListupSubscribe （未完成）
func ListupSubscribe(r *http.Request, pAccessToken *oauth2.Token, pPayload *SubscriptionListResponse) error {
	// WebHook エンドポイント
	pEndPoint := ""
	pRequest, pError := http.NewRequest("GET", pEndPoint, nil)
	if pError != nil {
		return fmt.Errorf("http.NewRequest(): %s", pError.Error())
	}
	pRequest.Header.Set("Authorization", "Bearer "+pAccessToken.AccessToken)
	pRequest.Header.Set("Content-Type", "application/json")

	// HTTPリクエストを送信
	pClient := &http.Client{Timeout: 10 * time.Second}
	pResponse, pError := pClient.Do(pRequest)
	if pError != nil {
		return fmt.Errorf("http.Client().Do(): %s", pError.Error())
	}
	defer pResponse.Body.Close()

	if pResponse.StatusCode != http.StatusOK {
		// エラーレスポンスのボディを読み取って詳細を出力できます
		bodyBytes, _ := io.ReadAll(pResponse.Body)
		return fmt.Errorf("ListupSubscribe() ==> StatusCode(%d) %s", pResponse.StatusCode, string(bodyBytes))
	}

	if pError := json.NewDecoder(pResponse.Body).Decode(pPayload); pError != nil {
		return fmt.Errorf("ListupSubscribe(): json.NewDecoder(): %s", pError.Error())
	}
	log.Print("pPayload: ")
	log.Println(pPayload)

	return nil
}
