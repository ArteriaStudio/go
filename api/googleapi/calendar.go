package googleapi

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/oauth2"
)

// CalendarstEntry はカレンダーリスト内の個々のカレンダー情報（エントリ）を表します。
type CalendarstEntry struct {
	// ID はカレンダーの一意な識別子（メールアドレス形式が多い）
	ID string `json:"id"`

	// Summary はユーザーが設定したカレンダーの名前
	Summary string `json:"summary"`

	// TimeZone はカレンダーのタイムゾーン
	TimeZone string `json:"timeZone"`

	// AccessRole はこのカレンダーに対するユーザーのアクセスレベル（例: owner, reader, writer, etc.）
	AccessRole string `json:"accessRole"`

	// Primary はこのカレンダーがユーザーのプライマリ（メイン）カレンダーであるかどうか
	Primary bool `json:"primary"`

	// 他にも多くのフィールドがありますが、主要なものを抜粋
}

// CalendarsContainer 複数のカレンダー情報のコンテナ
type CalendarsContainer struct {
	Kind          string             `json:"kind"`          // Kind はリソースのタイプ（例: "calendar#calendarList"）
	Etag          string             `json:"etag"`          // Etag はリソースのバージョンを表す不透明なトークン
	NextPageToken string             `json:"nextPageToken"` // NextPageToken は結果がページ分割されている場合に次のページのトークン
	Items         []*CalendarstEntry `json:"items"`         // Items はカレンダーエントリの配列
}

// CalendarInfo カレンダー情報
type CalendarInfo struct {
	ID                  string `json:"ID"`
	Name                string `json:"Name"`
	Color               string `json:"Color"`
	HexColor            string `json:"hexColor"`
	GroupClassId        string `json:"groupClassId"`
	IsDefaultCalendar   bool   `json:"isDefaultCalendar"`
	ChangeKey           string `json:"changeKey"`
	CanShare            bool   `json:"canShare"`
	CanViewPrivateItems bool   `json:"canViewPrivateItems"`
	CanEdit             bool   `json:"canEdit"`
	IsTallyingResponses bool   `json:"isTallyingResponses"`
	IsRemovable         bool   `json:"isRemovable"`
	Owner               struct {
		Name    string `json:"Name"`
		Address string `json:"Addres"`
	}
}

type CalendarDateTime struct {
	DateTime string `json:"dateTime"`
	TimeZone string `json:"timeZone"`
}

type CalendarLocation struct {
	DisplayName  string `json:"displayName"`
	LocationType string `json:"locationType"`
	UniqueId     string `json:"uniqueId"`
	UniqueIdType string `json:"uniqueIdType"`
}

type CalendarEmailAddress struct {
	Name    string `json:"Name"`
	Address string `json:"address"`
}

// 予定参加者の状態
type CalendarAttendStatus struct {
	Response   string    `json:"response"`
	StatusTime time.Time `json:"time"`
}

type CalendarAtendees struct {
	AtendeesType string `json:"type"`
	Status       CalendarAttendStatus
	Address      CalendarEmailAddress
}

// EventResource は Google Calendar API の単一のイベントを表す構造体です。
type EventResource struct {
	// イベントを一意に識別するID。これが eventId に相当します。
	ID string `json:"id"`

	// イベントの概要（タイトル）。
	Summary string `json:"summary"`

	// イベントの説明文。
	Description string `json:"description,omitempty"`

	// イベントのステータス（例: "confirmed", "tentative", "cancelled"）。
	Status string `json:"status"`

	// イベントの場所。
	Location string `json:"location,omitempty"`

	// イベントの最終更新日時 (RFC3339形式)。
	Updated time.Time `json:"updated"`

	// イベント作成日時 (RFC3339形式)。
	Created time.Time `json:"created"`

	// イベントの主催者。
	Organizer *Organizer `json:"organizer"`

	// イベントの参加者リスト。
	Attendees []Attendee `json:"attendees,omitempty"`

	// イベントの開始時刻と終了時刻。
	Start *EventDateTime `json:"start"`
	End   *EventDateTime `json:"end"`
}

// Organizer はイベントの主催者情報を保持します。
type Organizer struct {
	Email       string `json:"email"`
	DisplayName string `json:"displayName,omitempty"`
	Self        bool   `json:"self,omitempty"` // 主催者がリクエストユーザー自身かどうか
}

// Attendee はイベントの参加者情報を保持します。
type Attendee struct {
	Email          string `json:"email"`
	DisplayName    string `json:"displayName,omitempty"`
	Self           bool   `json:"self,omitempty"` // 参加者がリクエストユーザー自身かどうか
	ResponseStatus string `json:"responseStatus"` // 応答ステータス (例: "accepted", "tentative", "declined")
}

// EventDateTime はイベントの開始または終了時刻を保持します。
// 終日イベントと時刻指定イベントに対応できるように定義します。
type EventDateTime struct {
	// 時刻指定イベントの場合のRFC3339形式の時刻。
	DateTime time.Time `json:"dateTime,omitempty"`

	// 終日イベントの場合の日付（例: "2025-10-21"）。
	Date string `json:"date,omitempty"`

	// タイムゾーン（例: "Asia/Tokyo"）。
	TimeZone string `json:"timeZone,omitempty"`
}

// EventsListResponse は events.list メソッドの最上位レスポンス構造体です。
type EventsListResponse struct {
	Kind string `json:"kind"`
	Etag string `json:"etag"`

	// 次のページがある場合のトークン。
	NextPageToken string `json:"nextPageToken,omitempty"`

	// 次回の増分同期に使用するトークン。
	NextSyncToken string `json:"nextSyncToken,omitempty"`

	// 取得されたイベントのリスト。
	Items []*EventResource `json:"items"`
}

/*
type CalendarContent struct {
	Content     string `json:"content"`
	ContentType string `json:"contentType"`
}
*/

// GetCalendar カレンダー情報を取得
func GetCalendar(r *http.Request, pAccessToken *oauth2.Token, pCalenderInfo *CalendarInfo) error {
	// Access Tokenを使用してユーザー情報を取得
	pEndPoint := "https://www.googleapis.com/calendar/v3/calendars/primary"
	pRequest, pError := http.NewRequest("GET", pEndPoint, nil)
	if pError != nil {
		return fmt.Errorf("GetCalendar(): http.NewRequest(): %s", pError.Error())
	}
	pRequest.Header.Set("Authorization", "Bearer "+pAccessToken.AccessToken)
	pRequest.Header.Set("Accept", "application/json")

	pClient := &http.Client{Timeout: 10 * time.Second}
	pResponse, pError := pClient.Do(pRequest)
	if pError != nil {
		return fmt.Errorf("GetCalendar(): http.Client().Do(): %s", pError.Error())
	}
	defer pResponse.Body.Close()

	if pResponse.StatusCode != http.StatusOK {
		// エラーレスポンスのボディを読み取って詳細を出力できます
		return fmt.Errorf("GetCalendar(): %d", pResponse.StatusCode)
	}

	if pError := json.NewDecoder(pResponse.Body).Decode(pCalenderInfo); pError != nil {
		return fmt.Errorf("GetCalendar(): json.NewDecoder(): %s", pError.Error())
	}

	return nil
}

// GetCalendars カレンダーリストを取得
func GetCalendars(r *http.Request, pAccessToken *oauth2.Token, pCalendersContainer *CalendarsContainer) error {
	// Access Tokenを使用してユーザー情報を取得
	pEndPoint := "https://www.googleapis.com/calendar/v3/users/me/calendarList"
	pRequest, pError := http.NewRequest("GET", pEndPoint, nil)
	if pError != nil {
		return fmt.Errorf("GetCalendars(): http.NewRequest(): %s", pError.Error())
	}
	pRequest.Header.Set("Authorization", "Bearer "+pAccessToken.AccessToken)
	pRequest.Header.Set("Accept", "application/json")

	pClient := &http.Client{Timeout: 10 * time.Second}
	pResponse, pError := pClient.Do(pRequest)
	if pError != nil {
		return fmt.Errorf("GetCalendars(): http.Client().Do(): %s", pError.Error())
	}
	defer pResponse.Body.Close()

	if pResponse.StatusCode != http.StatusOK {
		// エラーレスポンスのボディを読み取って詳細を出力できます
		pBytes, _ := io.ReadAll(pResponse.Body)
		log.Print("GoogleAPI.GetCalendars(): Response ==> ")
		log.Println(string(pBytes))
		return fmt.Errorf("GetCalendars(): %d", pResponse.StatusCode)
	}

	if pError := json.NewDecoder(pResponse.Body).Decode(pCalendersContainer); pError != nil {
		return fmt.Errorf("GetCalendars(): json.NewDecoder(): %s", pError.Error())
	}

	return nil
}

// GetCalendarEvents イベントを列挙
func GetCalendarEvents(r *http.Request, pAccessToken *oauth2.Token, pCalendarId string, pCalendarEventInfo *EventsListResponse) error {
	// Access Tokenを使用してユーザー情報を取得
	if pCalendarId == "" {
		pCalendarId = "primary"
	}

	// URLとクエリパラメータを安全に構築
	u, err := url.Parse("https://www.googleapis.com/calendar/v3/calendars/" + pCalendarId + "/events")
	if err != nil {
		return fmt.Errorf("URL解析に失敗しました: %w", err)
	}

	q := u.Query()
	pUpdatedMin := time.Now().AddDate(0, 0, -1).Format(time.RFC3339)
	q.Set("updatedMin", pUpdatedMin)
	q.Set("showDeleted", "true")
	u.RawQuery = q.Encode()

	pEndPoint := u.String()
	//	pEndPoint := "https://www.googleapis.com/calendar/v3/calendars/" + pCalendarId + "/events" + pParam
	pRequest, pError := http.NewRequest("GET", pEndPoint, nil)
	if pError != nil {
		return fmt.Errorf("GetCalendarEvents(): http.NewRequest(): %s", pError.Error())
	}
	pRequest.Header.Set("Authorization", "Bearer "+pAccessToken.AccessToken)
	pRequest.Header.Set("Accept", "application/json")

	pClient := &http.Client{Timeout: 10 * time.Second}
	pResponse, pError := pClient.Do(pRequest)
	if pError != nil {
		return fmt.Errorf("GetCalendarEvents(): http.Client().Do(): %s", pError.Error())
	}
	defer pResponse.Body.Close()

	/*
		pBytes, _ := io.ReadAll(pResponse.Body)
		log.Print("GoogleAPI.GetCalendarEvent(): ")
		log.Println(string(pBytes))
	*/

	if pResponse.StatusCode != http.StatusOK {
		// エラーレスポンスのボディを読み取って詳細を出力できます
		return fmt.Errorf("GetCalendarEvents(): %d", pResponse.StatusCode)
	}
	if pError := json.NewDecoder(pResponse.Body).Decode(pCalendarEventInfo); pError != nil {
		return fmt.Errorf("GetCalendarEvents(): json.NewDecoder(): %s", pError.Error())
	}

	return nil
}

// GetCalendarEvent イベントを列挙
func GetCalendarEvent(r *http.Request, pAccessToken *oauth2.Token, pCalendarId string, pEventId string, pCalenderEventInfo *EventsListResponse) error {
	// Access Tokenを使用してユーザー情報を取得
	if pCalendarId == "" {
		pCalendarId = "primary"
	}
	pEndPoint := "https://www.googleapis.com/calendar/v3/calendars/" + pCalendarId + "/events/" + pEventId
	pRequest, pError := http.NewRequest("GET", pEndPoint, nil)
	if pError != nil {
		return fmt.Errorf("GetCalendarEvent(): http.NewRequest(): %s", pError.Error())
	}
	pRequest.Header.Set("Authorization", "Bearer "+pAccessToken.AccessToken)
	pRequest.Header.Set("Accept", "application/json")

	pClient := &http.Client{Timeout: 10 * time.Second}
	pResponse, pError := pClient.Do(pRequest)
	if pError != nil {
		return fmt.Errorf("GetCalendarEvent(): http.Client().Do(): %s", pError.Error())
	}
	defer pResponse.Body.Close()

	if pResponse.StatusCode != http.StatusOK {
		// エラーレスポンスのボディを読み取って詳細を出力できます
		return fmt.Errorf("GetCalendarEvent(): %d", pResponse.StatusCode)
	}
	if pError := json.NewDecoder(pResponse.Body).Decode(pCalenderEventInfo); pError != nil {
		return fmt.Errorf("GetCalendarEvent(): json.NewDecoder(): %s", pError.Error())
	}

	return nil
}
