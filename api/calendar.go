package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"arteria-s.net/entraapi"
	"arteria-s.net/googleapi"
	"golang.org/x/oauth2"
)

// Calendar カレンダーコンテキスト
type Calendar struct {
}

// HandlerListupGoogleCalendar Entraのカレンダーを一覧
func (pContext *Calendar) HandlerListupGoogleCalendar(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	var pCalendarsContainer googleapi.CalendarsContainer
	pError := pContext.PrepareListupGoogleCalendar(r, &c.pG.pAToken, &pCalendarsContainer)
	if pError != nil {
		log.Println(pError.Error())
		pURL := "/"
		http.Redirect(w, r, pURL, http.StatusTemporaryRedirect)
	} else {
		pContext.WriteListupGoogleCalendar(w, r, c, &pCalendarsContainer)
	}
}

// HandlerListupEntraCalendar Googleのカレンダーを一覧
func (pContext *Calendar) HandlerListupEntraCalendar(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	var pCalendarsContainer entraapi.CalendarsContainer
	pError := pContext.PrepareListupEntraCalendar(r, &c.pE.pAToken, &pCalendarsContainer)
	if pError != nil {
		log.Println(pError.Error())
		pURL := "/"
		http.Redirect(w, r, pURL, http.StatusTemporaryRedirect)
	} else {
		pContext.WriteListupEntraCalendar(w, r, c, &pCalendarsContainer)
	}
}

// HandlerDuplicateGoogleCalendar GoogleのカレンダーをEntraに複製
func (pContext *Calendar) HandlerDuplicateGoogleCalendar(w http.ResponseWriter, r *http.Request, c *FunctionContext) {

}

// HandlerDuplicateEntraCalendar EntraのカレンダーをGoogleに複製
func (pContext *Calendar) HandlerDuplicateEntraCalendar(w http.ResponseWriter, r *http.Request, c *FunctionContext) {

}

// PrepareListupGoogleCalendar
func (pContext *Calendar) PrepareListupGoogleCalendar(r *http.Request, pAccessToken *oauth2.Token, pCalendarsContainer *googleapi.CalendarsContainer) error {
	pError := googleapi.GetCalendars(r, pAccessToken, pCalendarsContainer)
	return pError
}

// PrepareListupEntraCalendar カレンダー一覧を獲得
func (pContext *Calendar) PrepareListupEntraCalendar(r *http.Request, pAccessToken *oauth2.Token, pCalendarsContainer *entraapi.CalendarsContainer) error {
	pError := entraapi.GetCalendars(r, pAccessToken, pCalendarsContainer)
	return pError
}

// WriteListupGoogleCalendar
func (pContext *Calendar) WriteListupGoogleCalendar(w http.ResponseWriter, r *http.Request, c *FunctionContext, pCalendaersContainer *googleapi.CalendarsContainer) {
	// ページを出力
	WriteHtmlHeaders(w, pApplicationName, http.StatusOK)

	fmt.Fprintln(w, "<body>")
	fmt.Fprintln(w, "<h2>カレンダーの一覧</h2>")
	fmt.Fprintln(w, "<p>"+time.Now().Format(time.RFC1123Z)+"<br /></p>")

	fmt.Fprintln(w, "<hr />")
	fmt.Fprintf(w, `<div class="header">`)
	fmt.Fprintln(w, "<p>")
	fmt.Fprintln(w, "<a href='/'>戻る</a>")
	fmt.Fprintln(w, "</p>")
	fmt.Fprintf(w, `</div>`)

	fmt.Fprintln(w, "<p>")
	for _, pCalendar := range pCalendaersContainer.Items {
		fmt.Fprintf(w, "Name: %s, IsDefault=%t,<br />", pCalendar.Summary, pCalendar.Primary)
	}

	fmt.Fprintln(w, "</p>")

	fmt.Fprintln(w, "<hr />")
	fmt.Fprintln(w, `<footer>`)
	fmt.Fprintln(w, "<p>Copyright 2025 Arteria Studio, All right reserved. </p>")
	fmt.Fprintln(w, "</footer>")

	fmt.Fprintln(w, "</body>")
}

// WriteListupEntraCalendar
func (pContext *Calendar) WriteListupEntraCalendar(w http.ResponseWriter, r *http.Request, c *FunctionContext, pCalendaersContainer *entraapi.CalendarsContainer) {
	// ページを出力
	WriteHtmlHeaders(w, pApplicationName, http.StatusOK)

	fmt.Fprintln(w, "<body>")
	fmt.Fprintln(w, "<h2>カレンダーの一覧</h2>")
	fmt.Fprintln(w, "<p>"+time.Now().Format(time.RFC1123Z)+"<br /></p>")

	fmt.Fprintln(w, "<hr />")
	fmt.Fprintf(w, `<div class="header">`)
	fmt.Fprintln(w, "<p>")
	fmt.Fprintln(w, "<a href='/'>戻る</a>")
	fmt.Fprintln(w, "</p>")
	fmt.Fprintf(w, `</div>`)

	fmt.Fprintln(w, "<p>")
	for _, pCalendar := range pCalendaersContainer.Value {
		fmt.Fprintf(w, "Name: %s, IsDefault=%t, Owner=%s(%s) Hex=%s(%s) Edit=%v<br />", pCalendar.Name, pCalendar.IsDefaultCalendar, pCalendar.Owner.Name, pCalendar.Owner.Address, pCalendar.Color, pCalendar.HexColor, pCalendar.CanEdit)
	}

	fmt.Fprintln(w, "</p>")

	fmt.Fprintln(w, "<hr />")
	fmt.Fprintln(w, `<footer>`)
	fmt.Fprintln(w, "<p>Copyright 2025 Arteria Studio, All right reserved. </p>")
	fmt.Fprintln(w, "</footer>")

	fmt.Fprintln(w, "</body>")
}
