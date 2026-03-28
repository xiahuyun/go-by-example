package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	bolt "go.etcd.io/bbolt"
)

var bucketName = []byte("users")

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Score int    `json:"score"`
}

func userKey(id int) []byte {
	return []byte(strconv.Itoa(id))
}

func upsertUser(db *bolt.DB, user User) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b == nil {
			return errors.New("bucket users does not exist")
		}

		data, err := json.Marshal(user)
		if err != nil {
			return fmt.Errorf("marshal user failed: %w", err)
		}

		return b.Put(userKey(user.ID), data)
	})
}

func getUser(db *bolt.DB, id int) (User, bool, error) {
	var (
		user  User
		found bool
	)

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b == nil {
			return errors.New("bucket users does not exist")
		}

		v := b.Get(userKey(id))
		if v == nil {
			return nil
		}

		if err := json.Unmarshal(v, &user); err != nil {
			return fmt.Errorf("unmarshal user failed: %w", err)
		}
		found = true
		return nil
	})

	return user, found, err
}

func deleteUser(db *bolt.DB, id int) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b == nil {
			return errors.New("bucket users does not exist")
		}
		return b.Delete(userKey(id))
	})
}

func printAllUsers(db *bolt.DB) error {
	return db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b == nil {
			return errors.New("bucket users does not exist")
		}

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var user User
			if err := json.Unmarshal(v, &user); err != nil {
				return fmt.Errorf("unmarshal user failed: %w", err)
			}
			fmt.Printf("Key=%s User=%+v\n", string(k), user)
		}
		return nil
	})
}

func main() {
	db, err := bolt.Open("demo.db", 0o600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatalf("open bbolt db failed: %v", err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		return err
	})
	if err != nil {
		log.Fatalf("init bucket failed: %v", err)
	}

	fmt.Println("== Create: 插入数据 ==")
	if err := upsertUser(db, User{ID: 1001, Name: "Alice", Score: 88}); err != nil {
		log.Fatal(err)
	}
	if err := upsertUser(db, User{ID: 1002, Name: "Bob", Score: 76}); err != nil {
		log.Fatal(err)
	}
	if err := upsertUser(db, User{ID: 1003, Name: "Cindy", Score: 93}); err != nil {
		log.Fatal(err)
	}
	if err := printAllUsers(db); err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n== Read: 按 ID 查询 ==")
	if user, ok, err := getUser(db, 1002); err != nil {
		log.Fatal(err)
	} else if ok {
		fmt.Printf("查询成功: %+v\n", user)
	} else {
		fmt.Println("查询失败: ID=1002 不存在")
	}

	fmt.Println("\n== Update: 更新 ID=1002 的分数 ==")
	if err := upsertUser(db, User{ID: 1002, Name: "Bob", Score: 84}); err != nil {
		log.Fatal(err)
	}
	if user, ok, err := getUser(db, 1002); err != nil {
		log.Fatal(err)
	} else if ok {
		fmt.Printf("更新后: %+v\n", user)
	}

	fmt.Println("\n== Delete: 删除 ID=1001 ==")
	if err := deleteUser(db, 1001); err != nil {
		log.Fatal(err)
	}
	fmt.Println("删除成功: ID=1001")

	fmt.Println("\n== Read: 删除后再次查询 ID=1001 ==")
	if user, ok, err := getUser(db, 1001); err != nil {
		log.Fatal(err)
	} else if ok {
		fmt.Printf("意外命中: %+v\n", user)
	} else {
		fmt.Println("查询结果: ID=1001 不存在")
	}

	fmt.Println("\n== 当前全部数据 ==")
	if err := printAllUsers(db); err != nil {
		log.Fatal(err)
	}
}
