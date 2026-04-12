package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"arteria-s.net/entraapi"
	"arteria-s.net/googleapi"
)

// PrepareListupSubscribe （未完成）
func PrepareListupSubscribeGoogle(r *http.Request, c *FunctionContext, pSubscribes *googleapi.SubscriptionListResponse) error {
	pError := googleapi.ListupSubscribe(r, &c.pG.pAToken, pSubscribes)

	return pError
}

// PrepareListupSubscribe
func PrepareListupSubscribeEntra(r *http.Request, c *FunctionContext, pSubscribes *entraapi.SubscriptionListResponse) error {
	pError := entraapi.ListupSubscribe(r, &c.pE.pAToken, pSubscribes)

	return pError
}

// HandlerListupGoogleSubscribe
func HandlerListupGoogleSubscribe(w http.ResponseWriter, r *http.Request, c *FunctionContext) {

}

// HandlerEntraListupSubscribe
func HandlerListupEntraSubscribe(w http.ResponseWriter, r *http.Request, c *FunctionContext) {
	// ページに出力する情報を収集
	var pSubscribes entraapi.SubscriptionListResponse
	pError := PrepareListupSubscribeEntra(r, c, &pSubscribes)
	if pError != nil {
		log.Println(pError.Error())
		pURL := "/"
		http.Redirect(w, r, pURL, http.StatusTemporaryRedirect)
	} else {
		// ページを出力
		WriteHtmlHeaders(w, pApplicationName, http.StatusOK)

		fmt.Fprintln(w, "<body>")
		fmt.Fprintln(w, "<h2>サブスクリプション一覧</h2>")
		fmt.Fprintln(w, "<p>"+time.Now().Format(time.RFC1123Z)+"<br /></p>")

		fmt.Fprintln(w, "<hr />")
		fmt.Fprintf(w, `<div class="header">`)
		fmt.Fprintln(w, "<p>")
		fmt.Fprintln(w, "<a href='/'>戻る</a>")
		fmt.Fprintln(w, "</p>")
		fmt.Fprintf(w, `</div>`)

		fmt.Fprintln(w, "<p>")
		for _, pSubscribe := range pSubscribes.Value {
			fmt.Fprintf(w, "SubscribeId: %s, ResourceId: %s, ExpireDate: %s<br />", pSubscribe.ID, pSubscribe.Resource, pSubscribe.ExpirationDateTime.Format(time.RFC3339))
		}

		fmt.Fprintln(w, "</p>")

		fmt.Fprintln(w, "<hr />")
		fmt.Fprintln(w, `<footer>`)
		fmt.Fprintln(w, "<p>Copyright 2025 Arteria Studio, All right reserved. </p>")
		fmt.Fprintln(w, "</footer>")

		fmt.Fprintln(w, "</body>")
	}
}
