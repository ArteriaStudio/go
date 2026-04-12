package entraapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

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

// CalendarEventInfo
type CalendarEventInfo struct {
	Id            string             `json:"id"`
	Subject       string             `json:"subject"`
	BodyPreview   string             `json:"bodyPreview"`
	HideAttendees bool               `json:"hideAttendees"`
	Start         CalendarDateTime   `json:"start"`
	End           CalendarDateTime   `json:"end"`
	Location      CalendarLocation   `json:"location"`
	Locations     []CalendarLocation `json:"locations"`
	Attendees     []CalendarAtendees `json:"attendees"`
	Organizer     struct {
		EmailAddress CalendarEmailAddress `json:"emailAddress"`
	} `json:"organizer"`
}

// CalendarEventContainer
type CalendarEventContainer struct {
	Value []CalendarEventInfo `json:"value"`
}

type CalendarContent struct {
	Content     string `json:"content"`
	ContentType string `json:"contentType"`
}

// GetCalendar カレンダー情報を取得
func GetCalendar(r *http.Request, pAccessToken *oauth2.Token, pCalendarInfo *CalendarInfo) error {
	// Access Tokenを使用してユーザー情報を取得
	pEndPoint := "https://graph.microsoft.com/v1.0/me/calendar"
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

	if pError := json.NewDecoder(pResponse.Body).Decode(pCalendarInfo); pError != nil {
		return fmt.Errorf("GetCalendar(): json.NewDecoder(): %s", pError.Error())
	}

	return nil
}

// GetCalendars カレンダー情報を取得
func GetCalendars(r *http.Request, pAccessToken *oauth2.Token, pCalendarInfo *CalendarInfo) error {
	// Access Tokenを使用してユーザー情報を取得
	pEndPoint := "https://graph.microsoft.com/v1.0/me/calendars"
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
		return fmt.Errorf("GetCalendars(): %d", pResponse.StatusCode)
	}

	if pError := json.NewDecoder(pResponse.Body).Decode(pCalendarInfo); pError != nil {
		return fmt.Errorf("GetCalendars(): json.NewDecoder(): %s", pError.Error())
	}

	return nil
}

// GetCalendarEvents イベントを列挙
func GetCalendarEvents(r *http.Request, pAccessToken *oauth2.Token, pResource string, pCalendarEventContainer *CalendarEventContainer) error {
	log.Printf("EntraAPI.GetCalendarEvents(): ENTRY ResourceId=[%s]\n", pResource)

	if pResource == "" {
		return fmt.Errorf("FAILED: GetCalendarEvents() pResource is null")
	}

	// Access Tokenを使用してユーザー情報を取得
	pEndPoint := "https://graph.microsoft.com/v1.0/" + pResource + "/calendar/events"
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
		log.Println("Body: " + string(pBytes))
	*/
	if pResponse.StatusCode != http.StatusOK {
		// エラーレスポンスのボディを読み取って詳細を出力できます
		return fmt.Errorf("GetCalendarEvents(): %d", pResponse.StatusCode)
	}
	if pError := json.NewDecoder(pResponse.Body).Decode(pCalendarEventContainer); pError != nil {
		return fmt.Errorf("GetCalendarEvents(): json.NewDecoder(): %s", pError.Error())
	}

	return nil
}

// GetCalendarEvent イベントを列挙
func GetCalendarEvent(r *http.Request, pAccessToken *oauth2.Token, pResource string, pCalendarEventInfo *CalendarEventInfo) error {
	// Access Tokenを使用してユーザー情報を取得
	pEndPoint := "https://graph.microsoft.com/v1.0/" + pResource + "?select=subject,body,bodyPreview,organizer,attendees,start,end,location,locations"
	pRequest, pError := http.NewRequest("GET", pEndPoint, nil)
	if pError != nil {
		return fmt.Errorf("GetCalendarEvent(): http.NewRequest(): %s", pError.Error())
	}
	pRequest.Header.Set("Authorization", "Bearer "+pAccessToken.AccessToken)
	pRequest.Header.Set("Accept", "application/json")
	pRequest.Header.Set("Prefer", `outlook.body-content-type="text/html"`)
	pRequest.Header.Set("Prefer", `outlook.timezone="Asia/Tokyo"`)

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
	if pError := json.NewDecoder(pResponse.Body).Decode(pCalendarEventInfo); pError != nil {
		return fmt.Errorf("GetCalendarEvent(): json.NewDecoder(): %s", pError.Error())
	}

	return nil
}
