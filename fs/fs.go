package fireStore

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"google.golang.org/api/iterator"
)

// 構造体でデータの型を定義
type Computer struct {
	Name       string `firestore:"name"`
	Ether      string `firestore:"ether"`
	WiFi       string `firestore:"wifi"`
	RemoteAddr string `firestore:"remoteaddr"`
}

// 　インスタンスを初期化
func init() {
	functions.HTTP("entryPoint", entryPoint)
}

// FirebaseプロジェクトIDを設定
var pProjectID = "spiral-44c1f"

// 　エントリーポイント
func entryPoint(w http.ResponseWriter, r *http.Request) {
	//　コンテキスト
	ctx := context.Background()

	// Firestoreクライアントを初期化
	pClient, err := firestore.NewClient(ctx, pProjectID)
	if err != nil {
		log.Fatalf("firestore.NewClient: %v", err)
	}
	defer pClient.Close()

	fmt.Fprintf(w, "Method: %s\n", r.Method)
	fmt.Fprintf(w, "URI: %s\n", r.RequestURI)
	fmt.Fprintf(w, "RemoteAddr: %s\n", r.RemoteAddr)

	if r.Method == "POST" {
		post(ctx, pClient, w, r)
	} else if r.Method == "GET" {
		get(ctx, pClient, w, r)
	}
}

// 　POSTメソッド
func post(ctx context.Context, pClient *firestore.Client, w http.ResponseWriter, r *http.Request) {

	// データを保存する
	collectionName := "computers"
	docID := "elise"
	pComputer := Computer{Name: "Alice Smith", Ether: "", WiFi: "", RemoteAddr: r.RemoteAddr}

	_, err := pClient.Collection(collectionName).Doc(docID).Set(ctx, pComputer)
	if err != nil {
		log.Fatalf("Failed to add computer: %v", err)
	}
	fmt.Fprintf(w, "Added computer: %v\n", pComputer)
}

// 　GETメソッド
func get(ctx context.Context, pClient *firestore.Client, w http.ResponseWriter, r *http.Request) {
	// データを取得する (ドキュメントIDを指定)
	collectionName := "computers"
	docID := "elise"

	doc, err := pClient.Collection(collectionName).Doc(docID).Get(ctx)
	if err != nil {
		log.Fatalf("Failed to get user: %v", err)
	}
	var retrievedData Computer
	doc.DataTo(&retrievedData)
	fmt.Printf("Retrieved user by ID:\n%+v\n", retrievedData)

	// データを取得する (クエリを使用)
	query := pClient.Collection(collectionName)
	iter := query.Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to iterate: %v", err)
		}
		var pComputer Computer
		doc.DataTo(&pComputer)
		fmt.Fprintf(w, "Retrieved computer:\n%+v\n", pComputer)

		v, err := json.Marshal(pComputer)
		if err != nil {
			fmt.Fprintln(w, "%w", err)
		} else {
			fmt.Printf("%s", string(v))
		}
	}
	fmt.Fprintf(w, "done.")
}
