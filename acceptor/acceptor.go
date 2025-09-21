// 　OAUTH2 認証コード受け取り（acceptor）
package acceptor

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

func init() {
	functions.HTTP("EntryPoint", EntryPoint)
}

// 　エントリーポイント
func EntryPoint(w http.ResponseWriter, r *http.Request) {
	//　コンテキスト
	//	pContext := context.Background()

	//　URIからリソース名を獲得
	pResults := strings.Split(r.RequestURI, "/")
	if len(pResults) < 3 {
		//　リソースが指定されていない。
		return
	} else {
		pCollection := pResults[1]
		pResourceId := pResults[2]
		if !IsExistCollection(pCollection) {
			return
		}
		fmt.Fprintf(w, "Collection: %s\n", pCollection)
		if pResourceId == "" {
			return
		}
		fmt.Fprintf(w, "ResourceId: %s\n", pResourceId)
	}
	fmt.Fprint(w, pResults)
}

// 　既知のコレクション名であるかを確認
func IsExistCollection(pCollection string) bool {
	if pCollection == "" {
		return false
	}
	if pCollection == "computers" {
		return true
	}
	if pCollection == "logging" {
		return true
	}
	if pCollection == "sessions" {
		return true
	}

	return false
}
