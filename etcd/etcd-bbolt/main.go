package main

import (
	"encoding/hex"
	"fmt"
	"log"

	"go.etcd.io/bbolt"
	"go.etcd.io/etcd/api/v3/mvccpb"
)

func main() {
	db, err := bbolt.Open("/Users/project/etcd/server/default.etcd/member/snap/db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 你从 bbolt keys 拿到的 revision（hex）
	keyHex := "00000000000000025f0000000000000000"

	// 转成二进制
	revKey, _ := hex.DecodeString(keyHex)
	fmt.Println("revKey:", string(revKey))

	db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("key"))
		val := b.Get(revKey)

		if val == nil {
			fmt.Println("not found")
			return nil
		}

		// 反序列化 protobuf
		kv := &mvccpb.KeyValue{}
		if err := kv.Unmarshal(val); err != nil {
			log.Fatal(err)
		}

		fmt.Println("user key  :", string(kv.Key))
		fmt.Println("user value:", string(kv.Value))
		fmt.Println("mod rev   :", kv.ModRevision)
		return nil
	})
}
