package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"arteria-s.net/entraapi"
)

// WriteBodyGoogleUserInfo
func WriteBodyGoogleUserInfo(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	fmt.Fprintf(w, `<div class="left-pane">`)
	fmt.Fprintf(w, `<div class="google-one-ring"><img class="profile-image" src="/profile/g"></div>`)
	fmt.Fprintf(w, `</div>`)

	fmt.Fprintf(w, `<div class="right-pane">`)
	fmt.Fprintf(w, "Name: %s<br />\n", c.pG.pUserInfo.Name)
	fmt.Fprintf(w, "Mail: %s<br />\n", c.pG.pUserInfo.Email)
	fmt.Fprintf(w, "Expire At: %s<br />\n", c.pG.pAToken.Expiry.Format(time.RFC1123Z))
	fmt.Fprintf(w, `</div>`)
}

// WriteBodyGoogleCalendar カレンダーを出力
func WriteBodyGoogleCalendar(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	//	var pCalendar googleapi.CalenderInfo
	//	googleapi.GetCalendar(r, &c.pE.pSToken, &pCalendar)
}

// WriteBodyEntraCalendar カレンダーを出力
func WriteBodyEntraCalendar(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	log.Println("WriteBodyEntraCalendar: ENTRY")
	var pCalendar entraapi.CalendarInfo
	var pCalendarsContainer entraapi.CalendarsContainer
	entraapi.GetCalendar(r, &c.pE.pAToken, &pCalendar)
	entraapi.GetCalendars(r, &c.pE.pAToken, &pCalendarsContainer)
	var pCalendarEventContainer entraapi.CalendarEventContainer
	pError := entraapi.GetCalendarEvents(r, &c.pE.pAToken, pCalendar.ID, &pCalendarEventContainer)
	if pError != nil {
		log.Println(pError.Error())
	} else {
		log.Println("---------")
		log.Println(pCalendarEventContainer)
		log.Println("---------")
	}
}

// WriteBodyEntraUserInfo
func WriteBodyEntraUserInfo(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	fmt.Fprintf(w, `<div class="left-pane">`)
	fmt.Fprintf(w, `<div class="google-one-ring"><img class="profile-image" src="/profile/e"></div>`)
	fmt.Fprintf(w, `</div>`)

	fmt.Fprintf(w, `<div class="right-pane">`)
	fmt.Fprintf(w, "Name: %s<br />\n", c.pE.pUserInfo.Name)
	fmt.Fprintf(w, "Mail: %s<br />\n", c.pE.pUserInfo.Email)
	fmt.Fprintf(w, "Expire At: %s<br />\n", c.pE.pAToken.Expiry.Format(time.RFC1123Z))
	fmt.Fprintf(w, `</div>`)
}

// WriteHtmlHeaders HTMLヘッダーをHTTP応答へ出力
func WriteHtmlHeaders(w http.ResponseWriter, pTitle string, statusCode int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)

	fmt.Fprintln(w, "<!DOCTYPE html>")
	fmt.Fprintln(w, `<html lang="ja"><head>`)
	fmt.Fprintf(w, "<title>%s</title>", pTitle)
	fmt.Fprintln(w, `<link rel="stylesheet" href="/style.css">`)
	//fmt.Fprintln(w, `<link rel="stylesheet" href="/google-style.css">`)
	fmt.Fprintln(w, "</head></html>")
}

// WriteSignupInfo サインアップ情報を出力
func WriteSignupInfo(w http.ResponseWriter, pUniqueId string, pSessionKey string, pMailad string, pPasswdRegist string, pPasswdVerify string, pVal string) {
	WriteHtmlHeaders(w, pApplicationName, http.StatusOK)

	fmt.Fprintln(w, "<body>")
	fmt.Fprintln(w, "<h2>サインアップ（arteria-s.net/auth/signup）</h2>")
	fmt.Fprintln(w, "<hr />")
	fmt.Fprintf(w, "<p>UniqueId: %s</p>\n", pUniqueId)
	fmt.Fprintf(w, "<p>SessionKey: %s</p>\n", pSessionKey)
	fmt.Fprintf(w, "<p>Mailad: %s</p>\n", pMailad)
	fmt.Fprintf(w, "<p>Passwd: %s, %s</p>\n", pPasswdRegist, pPasswdVerify)
	if pPasswdRegist == pPasswdVerify {
		fmt.Fprintf(w, "<p>HashValue: %s</p>\n", pVal)
	} else {
		fmt.Fprintf(w, "<p>HashValue: N/A(Mismatch)</p>\n")
	}
	fmt.Fprintln(w, "<hr />")
	fmt.Fprintln(w, `<p><a href="/">トップページに戻る。</p>`)
	fmt.Fprintln(w, "</body>")
}

// WriteLoginInfo ログイン情報を出力
func WriteLoginInfo(w http.ResponseWriter, pUniqueId string, pSessionKey string, pMailad string, pPasswd string, pVal string, nRows int) {
	WriteHtmlHeaders(w, pApplicationName, http.StatusOK)

	fmt.Fprintln(w, "<body>")
	fmt.Fprintln(w, "<h2>サインイン（arteria-s.net/auth/signin）</h2>")
	fmt.Fprintln(w, "<hr />")
	fmt.Fprintf(w, "<p>UniqueId: %s</p>\n", pUniqueId)
	fmt.Fprintf(w, "<p>SessionKey: %s</p>\n", pSessionKey)
	fmt.Fprintf(w, "<p>Mailad: %s</p>\n", pMailad)
	fmt.Fprintf(w, "<p>Passwd: %s</p>\n", pPasswd)
	fmt.Fprintf(w, "<p>HashValue: %s</p>\n", pVal)
	if nRows == 0 {
		fmt.Fprintf(w, "<p>アカウント状態: 未登録</p>\n")
	} else {
		fmt.Fprintf(w, "<p>アカウント状態: 登録済（%s, %s）</p>\n", pUniqueId, pMailad)
	}
	fmt.Fprintln(w, "<hr />")
	fmt.Fprintln(w, `<p><a href="/">トップページに戻る。</p>`)
	fmt.Fprintln(w, "</body>")
}

// WriteHome ホームページを出力
func WriteHome(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	fmt.Fprintln(w, "<body>")
	fmt.Fprintln(w, "<h2>ポータル</h2>")

	fmt.Fprintln(w, "<hr />")
	fmt.Fprintf(w, `<div class="header">`)
	fmt.Fprintln(w, "<p>")
	fmt.Fprintln(w, "UniqueId: "+c.pUniqueId+" ")
	fmt.Fprintln(w, "Name: "+c.pDisplayName+" ")
	fmt.Fprintln(w, "NowTime: "+time.Now().Format(time.RFC1123Z)+"<br />")
	fmt.Fprintln(w, "<a href='/auth/signout'>サインアウト</a>")
	fmt.Fprintln(w, "<a href='/unsubscribe/all'>サブスクリプションを解除</a>")
	fmt.Fprintln(w, "<a href='/subscribe/check'>サブスクリプションを検査</a>")
	fmt.Fprintln(w, "</p>")
	fmt.Fprintf(w, `</div>`)

	fmt.Fprintln(w, "<hr /><p>")
	fmt.Fprintf(w, `<div class="container">`)
	fmt.Fprintln(w, "Google<br />")
	if c.pG.pAToken.AccessToken != "" {
		WriteBodyGoogleUserInfo(w, r, c)
		fmt.Fprintln(w, "<span>")
		fmt.Fprintln(w, `<a href="/listup/google/calendar">カレンダーを一覧</a><br />`)
		fmt.Fprintln(w, `<a href="/duplicate/google/calendar">カレンダーを複製</a><br />`)
		fmt.Fprintln(w, `<a href="/synchronize/google/calendar">カレンダーを同期</a><br />`)
		fmt.Fprintln(w, `<a href="/listup/google/subscribe">サブスクリプションを一覧</a><br />`)
		fmt.Fprintln(w, `<a href="/subscribe/google">サブスクリプションを登録</a><br />`)
		fmt.Fprintln(w, `<a href="/perm/google/revoke">アクセス許可を破棄</a><br />`)
		fmt.Fprintln(w, "</span>")
	} else {
		fmt.Fprintln(w, "<span>")
		fmt.Fprintln(w, "<a href='/perm/google/grant'>アクセス許可を申請</a><br />")
		fmt.Fprintln(w, `<a href="/perm/google/revoke">アクセス許可を破棄</a><br />`)
		fmt.Fprintln(w, "</span>")
	}
	fmt.Fprintf(w, `</div>`)
	fmt.Fprintln(w, "</p>")

	fmt.Fprintln(w, "<hr /><p>")
	fmt.Fprintf(w, `<div class="container">`)
	fmt.Fprintln(w, "Entra<br />")
	if c.pE.pAToken.AccessToken != "" {
		WriteBodyEntraUserInfo(w, r, c)
		fmt.Fprintln(w, "<span>")
		fmt.Fprintln(w, `<a href="/listup/entra/calendar">カレンダーを一覧</a><br />`)
		fmt.Fprintln(w, `<a href="/duplicate/entra/calendar">カレンダーを複製</a><br />`)
		fmt.Fprintln(w, `<a href="/synchronize/entra/calendar">カレンダーを同期</a><br />`)
		fmt.Fprintln(w, `<a href="/listup/entra/subscribe">サブスクリプションを一覧</a><br />`)
		fmt.Fprintln(w, `<a href="/subscribe/entra">サブスクリプションを登録</a><br />`)
		fmt.Fprintln(w, `<a href="/perm/entra/revoke">アクセス許可を破棄</a><br />`)
		fmt.Fprintln(w, "</span>")
	} else {
		fmt.Fprintln(w, "<span>")
		fmt.Fprintln(w, "<a href='/perm/entra/grant'>アクセス許可を申請</a><br />")
		fmt.Fprintln(w, `<a href="/perm/entra/revoke">アクセス許可を破棄</a><br />`)
		fmt.Fprintln(w, "</span>")
	}
	fmt.Fprintf(w, `</div>`)
	fmt.Fprintln(w, "</p>")

	fmt.Fprintln(w, "<hr />")
	fmt.Fprintln(w, `<footer>`)
	fmt.Fprintln(w, "<p>Copyright 2025 Arteria Studio, All right reserved. </p>")
	fmt.Fprintln(w, "</footer>")

	fmt.Fprintln(w, "</body>")
}
