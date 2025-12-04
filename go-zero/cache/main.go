package main

import (
	"fmt"
	"github.com/zeromicro/go-zero/core/collection"
	"time"
)

func main() {
	cache, err := collection.NewCache(1*time.Minute, collection.WithLimit(1000))
	if err != nil {
		panic(err)
	}

	cache.Take("name", func() (interface{}, error) {
		time.Sleep(1 * time.Second)
		return "xia huyun", nil
	})

	cache.Set("nickname", "huyun")

	if v, ok := cache.Get("nickname"); ok {
		fmt.Println(v)
	}
	if v, ok := cache.Get("name"); ok {
		fmt.Println(v)
	}
}
