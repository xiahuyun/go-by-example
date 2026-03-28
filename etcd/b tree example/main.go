package main

import (
	"fmt"

	"github.com/google/btree"
)

type User struct {
	ID    int
	Name  string
	Score int
}

type UserItem struct {
	Key   int
	Value User
}

func (u UserItem) Less(than btree.Item) bool {
	return u.Key < than.(UserItem).Key
}

func upsertUser(tree *btree.BTree, user User) {
	tree.ReplaceOrInsert(UserItem{
		Key:   user.ID,
		Value: user,
	})
}

func getUser(tree *btree.BTree, id int) (User, bool) {
	item := tree.Get(UserItem{Key: id})
	if item == nil {
		return User{}, false
	}
	return item.(UserItem).Value, true
}

func getUserV(tree *btree.BTree, user User) (User, bool) {
	item := tree.Get(UserItem{Value: user})
	if item == nil {
		return User{}, false
	}
	return item.(UserItem).Value, true
}

func deleteUser(tree *btree.BTree, id int) bool {
	item := tree.Delete(UserItem{Key: id})
	return item != nil
}

func printAll(tree *btree.BTree) {
	tree.Ascend(func(i btree.Item) bool {
		user := i.(UserItem).Value
		fmt.Printf("ID=%d Name=%s Score=%d\n", user.ID, user.Name, user.Score)
		return true
	})
}

func main() {
	tree := btree.New(3)

	fmt.Println("== Create: 插入数据 ==")
	upsertUser(tree, User{ID: 1001, Name: "Alice", Score: 88})
	upsertUser(tree, User{ID: 1002, Name: "Bob", Score: 76})
	upsertUser(tree, User{ID: 1003, Name: "Cindy", Score: 93})
	printAll(tree)

	fmt.Println("\n== Read: 按 ID 查询 ==")
	if user, ok := getUser(tree, 1002); ok {
		fmt.Printf("查询成功: %+v\n", user)
	} else {
		fmt.Println("查询失败: ID=1002 不存在")
	}

	fmt.Println("\n== Read: 按 用户名 查询 ==")
	if user, ok := getUserV(tree, User{ID: 1002, Name: "Bob", Score: 76}); ok {
		fmt.Printf("查询成功: %+v\n", user)
	} else {
		fmt.Println("查询失败: ID=1002 不存在")
	}

	fmt.Println("\n== Update: 更新 ID=1002 的分数 ==")
	upsertUser(tree, User{ID: 1002, Name: "Bob", Score: 84})
	if user, ok := getUser(tree, 1002); ok {
		fmt.Printf("更新后: %+v\n", user)
	}

	fmt.Println("\n== Delete: 删除 ID=1001 ==")
	if deleted := deleteUser(tree, 1001); deleted {
		fmt.Println("删除成功: ID=1001")
	} else {
		fmt.Println("删除失败: ID=1001 不存在")
	}

	fmt.Println("\n== Read: 删除后再次查询 ID=1001 ==")
	if user, ok := getUser(tree, 1001); ok {
		fmt.Printf("意外命中: %+v\n", user)
	} else {
		fmt.Println("查询结果: ID=1001 不存在")
	}

	fmt.Println("\n== 当前全部数据（有序） ==")
	printAll(tree)
}
