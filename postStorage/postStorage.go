package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func main() {

	pProjectID := "lumine-52a20"
	pBucketName := "run-sources-lumine-52a20-asia-northeast1"

	ListBuckets(pProjectID)
	ListBucketObjects(pProjectID, pBucketName)
	Save(pBucketName, "services/receiver/object1.data")
}

func CreateClient() (pClient *storage.Client) {
	ctx := context.Background()

	// Cloud Storageクライアントを生成
	pClient, err := storage.NewClient(ctx, option.WithCredentialsFile("D:/home/profiles/Keys/unitdebug_lumine-52a20-4d0519ae6f54.json"))
	if err != nil {
		log.Fatalf("storage.NewClient: %v", err)
	}
	return (pClient)
}

func Save(pBucketName string, pObjectName string) {
	ctx := context.Background()

	// Cloud Storageクライアントを生成
	client, err := storage.NewClient(ctx, option.WithCredentialsFile("D:/home/profiles/Keys/unitdebug_lumine-52a20-4d0519ae6f54.json"))
	if err != nil {
		log.Fatalf("storage.NewClient: %v", err)
	}
	defer client.Close()

	// 書き込む文字列
	content := "Hello, Cloud Storage!"

	// オブジェクトを作成し、書き込み
	wc := client.Bucket(pBucketName).Object(pObjectName).NewWriter(ctx)
	if _, err := io.WriteString(wc, content); err != nil {
		log.Fatalf("io.WriteString: %v", err)
	}
	if err := wc.Close(); err != nil {
		log.Fatalf("wc.Close: %v", err)
	}

	fmt.Printf("Blob %v written to %v\n", pBucketName, pObjectName)
}

func ListBuckets(pProjectID string) {
	ctx := context.Background()

	// Cloud Storageクライアントを生成
	client, err := storage.NewClient(ctx, option.WithCredentialsFile("D:/home/profiles/Keys/unitdebug_lumine-52a20-4d0519ae6f54.json"))
	if err != nil {
		log.Fatalf("storage.NewClient: %v", err)
	}
	defer client.Close()

	it := client.Buckets(ctx, pProjectID)
	for {
		battrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Printf("Buckets(%q): %s", pProjectID, err.Error())
			return
		}
		fmt.Println(battrs.Name)
	}
}

func ListBucketObjects(pProjectID string, pBucketName string) {
	ctx := context.Background()

	// Cloud Storageクライアントを生成
	client, err := storage.NewClient(ctx, option.WithCredentialsFile("D:/home/profiles/Keys/unitdebug_lumine-52a20-4d0519ae6f54.json"))
	if err != nil {
		log.Fatalf("storage.NewClient: %v", err)
	}
	defer client.Close()

	it := client.Bucket(pBucketName).Objects(ctx, nil)
	for {
		battrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Printf("Buckets(%q): %s", pProjectID, err.Error())
			return
		}
		fmt.Println(battrs.Name)
	}
}
