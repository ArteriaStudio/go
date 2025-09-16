package postFunction

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"google.golang.org/api/iterator"
)

type NIC struct {
	Name        string `firestore:"name"`
	HWAddr      string `firestore:"hwaddr"`
	Description string `firestore:"Description"`
}

// 構造体でデータの型を定義
type Computer struct {
	Name       string `firestore:"name"`
	Domain     string `firestore:"domain"`
	Ether      string `firestore:"ether"`
	WiFi       string `firestore:"wifi"`
	UniqueID   string `firestore:"UniqueID"`
	SerialNo   string `firestore:"SerialNo"`
	Adapters   []NIC
	RemoteAddr string `firestore:"remoteaddr"`
	Timestamp  string `firestore:"timestamp"`
}

type Session struct {
	ComputerName string `firestore:"ComputerName"`
	DomainName   string `firestore:"DomainName"`
	EventType    int32  `firestore:"EventType"`
	UserName     string `firestore:"UserName"`
	RemoteAddr   string `firestore:"remoteaddr"`
	Timestamp    string `firestore:"TimeStamp"`
}

// インスタンスを初期化
func init() {
	functions.HTTP("EntryPoint", EntryPoint)
}

// FirebaseプロジェクトIDを設定
var pProjectID = "spiral-44c1f"

// 　エントリーポイント
func EntryPoint(w http.ResponseWriter, r *http.Request) {
	//　コンテキスト
	pContext := context.Background()

	// Firestoreクライアントを初期化
	pClient, err := firestore.NewClient(pContext, pProjectID)
	if err != nil {
		log.Fatalf("firestore.NewClient: %v", err)
	}
	defer pClient.Close()

	fmt.Fprintf(w, "Method: %s\n", r.Method)
	fmt.Fprintf(w, "URI: %s\n", r.RequestURI)
	fmt.Fprintf(w, "RemoteAddr: %s\n", r.RemoteAddr)

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

		if r.Method == "POST" {
			post(w, r, pContext, pClient, pCollection, pResourceId)
		} else if r.Method == "GET" {
			get(w, r, pContext, pClient, pCollection, pResourceId)
		}
	}
}

// 　POSTメソッド
func post(w http.ResponseWriter, r *http.Request, pContext context.Context, pClient *firestore.Client, pCollection string, pResourceId string) {
	//　リクエストボディを入力する。
	pBytes, err := io.ReadAll(r.Body)
	if err != nil {
		//　期待した形式のリクエストボディなのでリクエストを無視
		fmt.Fprintln(w, "%w", err)
		return
	}
	pText := string(pBytes)
	fmt.Fprintf(w, "<<%s>>", pText)

	switch pCollection {
	case "computers":
		postComputers(w, r, pContext, pClient, pCollection, pResourceId, pBytes)
	case "sessions":
		postSessions(w, r, pContext, pClient, pCollection, pResourceId, pBytes)
	}
}

// postComputers
func postComputers(w http.ResponseWriter, r *http.Request, pContext context.Context, pClient *firestore.Client, pCollection string, pResourceId string, pBytes []byte) {
	// データを保存する
	collectionName := pCollection
	docID := pResourceId

	var pRequest Computer
	pError := json.Unmarshal(pBytes, &pRequest)
	if pError != nil {
		fmt.Fprintln(w, "Unmarshal: %w", pError)
		return
	} else {
		fmt.Fprintf(w, "body: %s\n", string(pBytes))
		fmt.Fprintf(w, "Name: %s\n", pRequest.Name)
		fmt.Fprintf(w, "Ether: %s\n", pRequest.Ether)
		fmt.Fprintf(w, "Wi-Fi: %s\n", pRequest.WiFi)
	}

	pComputer := Computer{Name: pRequest.Name, Domain: pRequest.Domain, Ether: pRequest.Ether, WiFi: pRequest.WiFi, UniqueID: pRequest.UniqueID, SerialNo: pRequest.SerialNo, RemoteAddr: r.RemoteAddr, Timestamp: time.Now().String()}
	pComputer.Adapters = append(pComputer.Adapters, pRequest.Adapters...)

	_, err := pClient.Collection(collectionName).Doc(docID).Set(pContext, pComputer)
	if err != nil {
		log.Fatalf("Failed to add computer: %v", err)
	}
	fmt.Fprintf(w, "Added computer: %v\n", pComputer)
}

// postSessions
func postSessions(w http.ResponseWriter, r *http.Request, pContext context.Context, pClient *firestore.Client, pCollection string, pResourceId string, pBytes []byte) {
	// データを保存する
	collectionName := pCollection
	docID := pResourceId

	var pRequest Session
	pError := json.Unmarshal(pBytes, &pRequest)
	if pError != nil {
		fmt.Fprintln(w, "Unmarshal: %w", pError)
		return
	} else {
		fmt.Fprintf(w, "body: %s\n", string(pBytes))
		fmt.Fprintf(w, "ComputerName: %s\n", pRequest.ComputerName)
		fmt.Fprintf(w, "DomainName: %s\n", pRequest.DomainName)
		fmt.Fprintf(w, "EventType: %d\n", pRequest.EventType)
		fmt.Fprintf(w, "UserName: %s\n", pRequest.UserName)
		fmt.Fprintf(w, "Timestamp: %s\n", pRequest.Timestamp)
	}

	pSession := Session{ComputerName: pRequest.ComputerName, DomainName: pRequest.DomainName, EventType: pRequest.EventType, UserName: pRequest.UserName, RemoteAddr: r.RemoteAddr, Timestamp: pRequest.Timestamp}

	timeStamp := pSession.Timestamp
	_, err := pClient.Collection(collectionName).Doc(docID).Collection("Activity").Doc(timeStamp).Set(pContext, pSession)
	//	_, err := pClient.Collection(collectionName).Doc(docID).Set(pContext, pSession)
	if err != nil {
		log.Fatalf("Failed to add sessions: %v", err)
	}
	fmt.Fprintf(w, "Added Session: %v\n", pSession)
}

// 　GETメソッド
func get(w http.ResponseWriter, r *http.Request, pContext context.Context, pClient *firestore.Client, pCollection string, pResourceId string) {
	// データを取得する (ドキュメントIDを指定)
	collectionName := pCollection
	docID := pResourceId

	doc, err := pClient.Collection(collectionName).Doc(docID).Get(pContext)
	if err != nil {
		log.Printf("Failed to get user: %v", err)
		return
	}
	var retrievedData Computer
	doc.DataTo(&retrievedData)

	// データを取得する (クエリを使用)
	query := pClient.Collection(collectionName)
	iter := query.Documents(pContext)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Failed to iterate: %v", err)
			return
		}
		var pComputer Computer
		doc.DataTo(&pComputer)
		fmt.Fprintf(w, "Retrieved computer:\n%+v\n", pComputer)

		v, err := json.Marshal(pComputer)
		if err != nil {
			fmt.Fprintln(w, "%w", err)
			return
		} else {
			fmt.Printf("%s", string(v))
		}
	}
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
