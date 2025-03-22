package main

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// 構造体でデータの型を定義
type User struct {
	Name string `firestore:"name"`
	Age  int    `firestore:"age"`
	City string `firestore:"city"`
}

type Computer struct {
	Name  string `firestore:"name"`
	Ether string `firestore:"ether"`
	WiFi  string `firestore:"wifi"`
}

func main() {
	pComputerName := "elise"

	ctx := context.Background()

	// FirebaseプロジェクトIDを設定
	projectID := "spiral-44c1f"

	// Firestoreクライアントを初期化
	client, err := firestore.NewClient(ctx, projectID, option.WithCredentialsFile("D:/home/profiles/Keys/spiral-44c1f-a76f129455f3.json"))
	if err != nil {
		log.Fatalf("firestore.NewClient: %v", err)
	}
	defer client.Close()

	// データを保存する
	collectionName := "computers"
	docID := pComputerName
	pComputer := Computer{Name: "Alice Smith", Ether: "00:00:00:00:00:00", WiFi: "00:00:00:00:00:01"}

	_, err = client.Collection(collectionName).Doc(docID).Set(ctx, pComputer)
	if err != nil {
		log.Fatalf("Failed to add user: %v", err)
	}
	fmt.Printf("Added user: %v\n", pComputer)

	// データを取得する (ドキュメントIDを指定)
	doc, err := client.Collection(collectionName).Doc(docID).Get(ctx)
	if err != nil {
		log.Fatalf("Failed to get user: %v", err)
	}
	var retrievedAlice Computer
	doc.DataTo(&retrievedAlice)
	fmt.Printf("Retrieved user by ID:\n%+v\n", retrievedAlice)

	// データを取得する (クエリを使用)
	query := client.Collection(collectionName)
	//	query := client.Collection(collectionName).Where("city", "==", "New York")
	iter := query.Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to iterate: %v", err)
		}
		var computer Computer
		doc.DataTo(&computer)
		//		fmt.Printf("Retrieved user by query (city=New York):\n%+v\n", computer)
		fmt.Printf("Retrieved computer:\n%+v\n", computer)
	}
}
