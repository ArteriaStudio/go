package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// 構造体でデータの型を定義

type NIC struct {
	Name   string `firestore:"name"`
	HWAddr string `firestore:"hwaddr"`
}

// 構造体でデータの型を定義
type Computer struct {
	Name       string `firestore:"name"`
	Domain     string `firestore:"domain"`
	Ether      string `firestore:"ether"`
	WiFi       string `firestore:"wifi"`
	Adapters   []NIC
	RemoteAddr string `firestore:"remoteaddr"`
	Timestamp  string `firestore:"timestamp"`
}

func main() {
	//	var pText1 = "{\"Name\":\"WAKABA\",\"Domain\":\"W\",\"Ether\":\"50:c2:e8:65:85:4c,b4:45:06:ac:58:07,\",\"WiFi\":\"52:c2:e8:65:85:4b,56:c2:e8:65:85:4b,50:c2:e8:65:85:4b,\",\"Adapters\":[{\"Name\":\"ローカル エリア接続* 1\",\"HWAddr\":\"52:c2:e8:65:85:4b\",},{\"Name\":\"ローカル エリア接続* 2\",\"HWAddr\":\"56:c2:e8:65:85:4b\",},{\"Name\":\"Wi-Fi\",\"HWAddr\":\"50:c2:e8:65:85:4b\",},{\"Name\":\"Bluetooth ネットワーク接続\",\"HWAddr\":\"50:c2:e8:65:85:4c\",},{\"Name\":\"イーサネット\",\"HWAddr\":\"b4:45:06:ac:58:07\",},]}"
	//	var pText0 = "{\"Name\":\"WAKABA\",\"Domain\":\"W\",\"Ether\":\"50:c2:e8:65:85:4c,b4:45:06:ac:58:07,\",\"WiFi\":\"52:c2:e8:65:85:4b,56:c2:e8:65:85:4b,50:c2:e8:65:85:4b,\",\"Adapters\":[{\"Name\":\"ローカル エリア接続* 1\",\"HWAddr\":\"52:c2:e8:65:85:4b\",},{\"Name\":\"ローカル エリア接続* 2\",\"HWAddr\":\"56:c2:e8:65:85:4b\",},{\"Name\":\"Wi-Fi\",\"HWAddr\":\"50:c2:e8:65:85:4b\",},{\"Name\":\"Bluetooth ネットワーク接続\",\"HWAddr\":\"50:c2:e8:65:85:4c\",},{\"Name\":\"イーサネット\",\"HWAddr\":\"b4:45:06:ac:58:07\",}]}"
	var pText0 = "{\"Name\":\"WAKABA\",\"Domain\":\"W\",\"Ether\":\"50:c2:e8:65:85:4c,b4:45:06:ac:58:07,\",\"WiFi\":\"52:c2:e8:65:85:4b,56:c2:e8:65:85:4b,50:c2:e8:65:85:4b,\",\"Adapters\":[{\"Name\":\"ローカル エリア接続* 1\",\"HWAddr\":\"52:c2:e8:65:85:4b\"},{\"Name\":\"ローカル エリア接続* 2\",\"HWAddr\":\"56:c2:e8:65:85:4b\"},{\"Name\":\"Wi-Fi\",\"HWAddr\":\"50:c2:e8:65:85:4b\"},{\"Name\":\"Bluetooth ネットワーク接続\",\"HWAddr\":\"50:c2:e8:65:85:4c\"},{\"Name\":\"イーサネット\",\"HWAddr\":\"b4:45:06:ac:58:07\"}]}"
	pBytes := []byte(pText0)
	var pRequest Computer
	pError := json.Unmarshal(pBytes, &pRequest)
	if pError != nil {
		fmt.Println("Unmarshal: %w", pError)
		return
	}

	pBytes2, pError := json.Marshal(pRequest)
	if pError != nil {
		fmt.Println("Unmarshal: %w", pError)
		return
	}
	fmt.Printf("Marshal... %s\n", string(pBytes2))

}

func main2() {
	pURI := "/computers/elise"
	pResults := strings.Split(pURI, "/")
	fmt.Println(pResults)

	pComputerName := "elise"

	ctx := context.Background()

	// FirebaseプロジェクトIDを設定
	projectID := "spiral-44c1f"

	// Firestoreクライアントを初期化
	client, err := firestore.NewClient(ctx, projectID, option.WithCredentialsFile("D:/home/profiles/Keys/spiral-44c1f-edfb90825b62.json"))
	if err != nil {
		log.Fatalf("firestore.NewClient: %v", err)
	}
	defer client.Close()

	// データを保存する
	collectionName := "computers"
	docID := pComputerName
	pComputer := Computer{Name: "Alice Smith", Ether: "00:00:00:00:00:00", WiFi: "00:00:00:00:00:01", Timestamp: time.Now().String()}

	var pAdapter NIC
	pAdapter.Name = "Adapter1"
	pAdapter.HWAddr = "11:22:11:22:11:22"
	pComputer.Adapters = append(pComputer.Adapters, pAdapter)

	pAdapter.Name = "Adapter2"
	pAdapter.HWAddr = "11:22:11:22:11:33"
	pComputer.Adapters = append(pComputer.Adapters, pAdapter)

	_, err = client.Collection(collectionName).Doc(docID).Set(ctx, pComputer)
	if err != nil {
		log.Fatalf("Failed to add user: %v", err)
	}
	fmt.Printf("Added user: %v\n", pComputer)

	v, err := json.Marshal(pComputer)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%s", string(v))
	}

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

		var pAdapter NIC
		pAdapter.Name = "Adapter1"
		pAdapter.HWAddr = "11:22:11:22:11:22"
		computer.Adapters = append(computer.Adapters, pAdapter)

		pAdapter.Name = "Adapter2"
		pAdapter.HWAddr = "11:22:11:22:11:33"
		computer.Adapters = append(computer.Adapters, pAdapter)

		v, err := json.Marshal(computer)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("%s", string(v))
		}
		fmt.Printf("Retrieved computer:\n%+v\n", computer)
	}
}
