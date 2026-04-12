// リスナー処理
package signin

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"arteria-s.net/entraapi"
	"arteria-s.net/googleapi"
	"arteria-s.net/postgres"
)

// NotificationPayload Microsoft Graph からの通知ペイロード全体に対応する構造体
type NotificationPayload struct {
	Value []Notification `json:"value"`
}

// Notification 個別の通知アイテムに対応する構造体 (必要なフィールドのみ)
type Notification struct {
	SubscriptionId     string `json:"subscriptionId"`
	ExpirationDateTime string `json:"subscriptionExpirationDateTime"`
	TenantID           string `json:"tenantId"`
	ClientState        string `json:"clientState"`
	LifecycleEvent     string `json:"lifecycleEvent"`
	ChangeType         string `json:"changeType"`
	Resource           string `json:"resource"`
}

// HandlerListenerURI リスナーハンドラ
func HandlerListenerURI(w http.ResponseWriter, r *http.Request, pURI string) {
	// 関数コンテキストを作成
	var c FunctionContext

	// データベースと接続
	pServerHostName := "localhost"
	pDatabaseName := "abook"
	c.pDatabase = postgres.Open(pServerHostName, pDatabaseName)
	defer postgres.Close(c.pDatabase)

	// エンドポイントのグローバルIPアドレスを獲得
	c.pClientIP = GetClientIP(r)

	// EntraID向けの共通フック処理（validationTokenがある場合は、固定の応答を返して終える。）
	pValidationToken := r.URL.Query().Get("validationToken")
	if pValidationToken != "" {
		// 検証リクエストの場合、validationTokenの値をそのまま返す
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(pValidationToken))
		log.Println("CATCH: Listener Validation. Response OK, in " + pURI)
		return
	}

	// セッション識別子にアクセストークンを紐付
	//PrepareAccessTokens(&c)

	// イベントハンドラを実行
	switch pURI {
	case "/listener/google":
		HandlerGoogleListener(w, r, &c)
	case "/listener/google/lifecycle":
		HandlerGoogleLifeCycleListener(w, r, &c)
	case "/listener/entra":
		HandlerEntraListener(w, r, &c)
	case "/listener/entra/lifecycle":
		HandlerEntraLifeCycleListener(w, r, &c)
	default:
	}
}

// HandlerGoogleLifeCycleListener
func HandlerGoogleLifeCycleListener(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	log.Println("HandlerGoogleLifeCycleListener(): ENTRY")

	var pNotifys NotificationPayload
	pError := json.NewDecoder(r.Body).Decode(&pNotifys)
	if pError != nil {
		log.Printf("JSON デコードエラー: %v", pError)
		return
	}
	for _, pNotify := range pNotifys.Value {
		log.Printf("変更タイプ: %s, リソース: %s, サブスクリプション ID: %s", pNotify.ChangeType, pNotify.Resource, pNotify.SubscriptionId)
		log.Println(pNotify)
		pLifecycleEvent := pNotify.LifecycleEvent
		pSubscribeId := pNotify.SubscriptionId
		pClientState := pNotify.ClientState
		log.Println("LifecycleEvent: " + pLifecycleEvent)
		log.Println("SubscribeId: " + pSubscribeId)
		log.Println("ClientState: " + pClientState)

		nRows, pError := postgres.LookupSubscribe(c.pDatabase, postgres.AuthorityEntra, pSubscribeId, pClientState, &c.pUniqueId, &c.pE.pId)
		if pError != nil {
			log.Printf("postgres.LookupSubscribe() %v\n", pError)
			continue
		}
		if nRows == 0 {
			continue
		}
		nRows, pError = postgres.LookupDelegateTokenById(c.pDatabase, postgres.AuthorityEntra, c.pUniqueId, &c.pE.pAToken)
		if pError != nil {
			log.Printf("LookupDelegateTokenById() is failed. %s\n", pError.Error())
			continue
		}
		// データベース更新などの処理
		if nRows == 0 {
			continue
		}
		pError = PrepareEntraToken(r, c.pDatabase, c.pUniqueId, true, &c.pE)
		if pError != nil {
			log.Printf("PrepareEntraToken() is failed. %s\n", pError.Error())
			continue
		}

		//
		pExpireTimeStamp := time.Now().UTC().Add(time.Minute * 45)
		pError = entraapi.SubscribeUpdate(r, &c.pE.pAToken, pExpireTimeStamp, pSubscribeId)
		if pError != nil {
			log.Println("ERROR: HandlerGoogleLifeCycleListener() " + pError.Error())
			continue
		}
		log.Printf("Sent Subscription update.SubscribeID=[%s]\n", pSubscribeId)
	}
}

// HandlerEntraLifeCycleListener
func HandlerEntraLifeCycleListener(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	var pNotifys NotificationPayload
	pError := json.NewDecoder(r.Body).Decode(&pNotifys)
	if pError != nil {
		log.Printf("ERROR: HandlerEntraLifeCycleListener() JSON デコードエラー: %v", pError)
		return
	}
	for _, pNotify := range pNotifys.Value {
		log.Printf("変更タイプ: %s, リソース: %s, サブスクリプション ID: %s", pNotify.ChangeType, pNotify.Resource, pNotify.SubscriptionId)

		pLifecycleEvent := pNotify.LifecycleEvent
		pSubscribeId := pNotify.SubscriptionId
		pClientState := pNotify.ClientState

		log.Println("LifecycleEvent: " + pLifecycleEvent)
		log.Println("SubscribeId: " + pSubscribeId)
		log.Println("ClientState: " + pClientState)

		nRows, pError := postgres.LookupSubscribe(c.pDatabase, postgres.AuthorityEntra, pSubscribeId, pClientState, &c.pUniqueId, &c.pE.pId)
		if pError != nil {
			log.Printf("ERROR: postgres.LookupSubscribe() %v\n", pError)
			continue
		}
		if nRows == 0 {
			continue
		}
		nRows, pError = postgres.LookupDelegateTokenById(c.pDatabase, postgres.AuthorityEntra, c.pUniqueId, &c.pE.pAToken)
		if pError != nil {
			log.Printf("ERROR: LookupDelegateTokenById() is failed. %s\n", pError.Error())
			continue
		}
		if nRows == 0 {
			continue
		}

		// データベース更新などの処理
		pError = PrepareEntraToken(r, c.pDatabase, c.pUniqueId, true, &c.pE)
		if pError != nil {
			log.Printf("ERROR: HandlerEntraLifeCycleListener::PrepareEntraToken(): %s", pError.Error())
		}
		//
		pExpireTimeStamp := time.Now().UTC().Add(time.Minute * 45)
		pError = entraapi.SubscribeUpdate(r, &c.pE.pAToken, pExpireTimeStamp, pSubscribeId)
		if pError != nil {
			log.Println("ERROR: HandlerEntraLifeCycleListener() " + pError.Error())
		} else {
			log.Printf("Sent Subscription update.SubscribeID=[%s]\n", pSubscribeId)
		}
	}
}

// HandlerGoogleListener リスナー
// Google Calendar APIからの変更通知は、Graph APIと異なり「更新された契機」のみが届く。
// Entraと異なりGoogle APIは変更履歴のdeltaを獲得する方法が提供されているので、チャネル（サブスクリプション）を維持する必要が少ない。
func HandlerGoogleListener(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	if r.Method != "POST" {
		// POSTメソッド以外はイベントを無視する。
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	pSubscribeId := r.Header.Get("X-Goog-Channel-Id")
	pChannelExpire := r.Header.Get("X-Goog-Channel-Expiration")
	pResourceState := r.Header.Get("X-Goog-Resource-State")
	pResourceUri := r.Header.Get("X-Goog-Resource-Uri")
	pMessageNo := r.Header.Get("X-Goog-Message-Number")
	pResourceId := r.Header.Get("X-Goog-Resource-Id")
	pChannelToken := r.Header.Get("X-Goog-Channel-Token")

	log.Printf("ChannelId: %s\n", pSubscribeId)
	log.Printf("ChannelExpire: %s\n", pChannelExpire)
	log.Printf("ResourceState: %s\n", pResourceState)
	log.Printf("ResourceUri: %s\n", pResourceUri)
	log.Printf("MessageNo: %s\n", pMessageNo)
	log.Printf("ResourceId: %s\n", pResourceId)
	log.Printf("Token: %s\n", pChannelToken)

	// チャネルトークンとサブスクリプション識別子が一致する情報の有無を確認
	nRows, pError := postgres.LookupSubscribe(c.pDatabase, postgres.AuthorityGoogle, pSubscribeId, pChannelToken, &c.pUniqueId, &c.pG.pId)
	if pError != nil {
		log.Printf("HandlerGoogleListener::postgres.LookupSubscribe() %v\n", pError)
		return
	}
	if nRows == 0 {
		return
	}
	nRows, pError = postgres.LookupDelegateTokenById(c.pDatabase, postgres.AuthorityGoogle, c.pUniqueId, &c.pG.pAToken)
	if pError != nil {
		log.Printf("LookupDelegateTokenById() is failed. %s\n", pError.Error())
		return
	}
	if nRows == 0 {
		return
	}

	log.Printf("UniqueId: %s\n", c.pUniqueId)
	log.Printf("AccountId: %s\n", c.pG.pId)
	log.Printf("AccessToken: %s\n", c.pG.pAToken.AccessToken)
	log.Printf("RefreshToken: %s\n", c.pG.pAToken.RefreshToken)
	pError = PrepareGoogleToken(r, c.pDatabase, c.pUniqueId, false, &c.pG)
	if pError != nil {
		log.Printf("PrepareGoogleToken() is failed. %s\n", pError.Error())
		return
	}

	var pCalenderEventInfo googleapi.EventsListResponse
	pCalendarId := "primary"
	pError = googleapi.GetCalendarEvents(r, &c.pG.pAToken, pCalendarId, &pCalenderEventInfo)
	if pError != nil {
		log.Println(pError.Error())
		return
	}

	log.Print("CalenderEventInfo: ")
	log.Println(pCalenderEventInfo)
	if len(pCalenderEventInfo.Items) > 0 {
		log.Println(pCalenderEventInfo.Items[0])
	}

	// 通知を受け付けたことをGraph APIに伝えるために 202 OK を返す
	w.WriteHeader(http.StatusAccepted) // または 200 OK
}

// HandlerEntraListener リスナー
// Microsoft Graph APIの場合、イベント変更通知のリソース識別子に変更されたイベントデータの識別子が送られてくる。
// 新規登録と更新はクエリーすればイベント情報を獲得できるが、削除の場合はクエリーしても応答にも削除されたイベント情報を獲得できない。
// → 削除を伝搬するためには事前に情報を保持している必要がある。
func HandlerEntraListener(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	if r.Method != "POST" {
		// POSTメソッド以外はイベントを無視する。
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	log.Println("HandlerEntraListener(): ENTRY")

	var pNotifys NotificationPayload
	pError := json.NewDecoder(r.Body).Decode(&pNotifys)
	if pError != nil {
		log.Printf("HandlerEntraListener(): JSON デコードエラー: %v", pError)
		return
	}

	// 変換されたデータを使ってビジネスロジックを実行
	for _, pNotify := range pNotifys.Value {
		pExpireTimeStamp := GetLocalTimeStampByRFC3399(pNotify.ExpirationDateTime)
		log.Printf("変更タイプ: %s 有効期限：%s\n", pNotify.ChangeType, pExpireTimeStamp)

		// ClientState の検証
		pClientState := pNotify.ClientState
		pSubscribeId := pNotify.SubscriptionId

		nRows, pError := postgres.LookupSubscribe(c.pDatabase, postgres.AuthorityEntra, pSubscribeId, pClientState, &c.pUniqueId, &c.pE.pId)
		if pError != nil {
			log.Printf("postgres.LookupSubscribe() %v\n", pError)
			continue
		}
		if nRows == 0 {
			log.Printf("postgres.LookupSubscribe() nRows=%d\n", nRows)
			continue
		}

		nRows, pError = postgres.LookupDelegateTokenById(c.pDatabase, postgres.AuthorityEntra, c.pUniqueId, &c.pE.pAToken)
		if pError != nil {
			log.Printf("LookupDelegateTokenById() is failed. %s\n", pError.Error())
			continue
		}

		// データベース更新などの処理
		if nRows == 0 {
			log.Printf("LookupDelegateTokenById() is nRows=%d\n", nRows)
			continue
		}
		pError = PrepareEntraToken(r, c.pDatabase, c.pUniqueId, false, &c.pE)
		if pError != nil {
			log.Printf("ERROR: PrepareEntraToken() %s\n", pError.Error())
			continue
		}

		// 変更通知が示すイベント情報を取得（ResouceId="Users/.../events/..."）
		var pCalendarEventInfo entraapi.CalendarEventInfo
		pError = entraapi.GetCalendarEvent(r, &c.pE.pAToken, pNotify.Resource, &pCalendarEventInfo)
		if pError != nil {
			log.Println(pError.Error())
			continue
		}
	}

	// 通知を受け付けたことをGraph APIに伝えるために 202 OK を返す
	w.WriteHeader(http.StatusAccepted) // または 200 OK
}
