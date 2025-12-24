package main

import (
	"fmt"
	"time"

	"github.com/go-zookeeper/zk"
)

func main() {
	servers := []string{"localhost:2181"}
	conn, _, err := zk.Connect(servers, time.Second*5)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	path := "/my_go_node"

	// 1. 先删除可能已存在的节点（忽略版本）
	if exists, _, _ := conn.Exists(path); exists {
		err = conn.Delete(path, -1) // -1 表示忽略版本检查
		if err == nil {
			fmt.Println("Deleted existing node")
		}
	}

	// 2. 创建节点
	data := []byte("Hello ZooKeeper from Go!")
	createdPath, err := conn.Create(path, data, 0, zk.WorldACL(zk.PermAll))
	if err != nil {
		fmt.Printf("Create failed: %v\n", err)
	} else {
		fmt.Printf("Node created: %s\n", createdPath)
	}

	// 3. 获取节点数据和最新状态
	currentData, currentStat, err := conn.Get(path)
	if err != nil {
		fmt.Printf("Get data failed: %v\n", err)
	} else {
		fmt.Printf("Node data: %s (Version: %d)\n", string(currentData), currentStat.Version)
	}

	// 4. 设置监听
	go watchNode(conn, path)

	// 5. 更新节点（使用最新版本号）
	newData := []byte("Updated data!")
	newStat, err := conn.Set(path, newData, currentStat.Version)
	if err != nil {
		fmt.Printf("Update failed: %v\n", err)
	} else {
		fmt.Printf("Node updated successfully. New version: %d\n", newStat.Version)
	}

	// 6. 短暂等待观察器触发
	time.Sleep(500 * time.Millisecond)

	// 7. 删除节点（使用更新后的最新版本号）
	err = conn.Delete(path, newStat.Version) // 使用更新操作返回的新版本号
	if err != nil {
		fmt.Printf("Delete failed: %v\n", err)
	} else {
		fmt.Println("Node deleted successfully")
	}

	// 等待最终观察事件
	time.Sleep(500 * time.Millisecond)
}

// 观察节点变化的函数
func watchNode(conn *zk.Conn, path string) {
	for {
		// 获取数据并设置观察
		_, _, eventCh, err := conn.GetW(path)
		if err != nil {
			fmt.Printf("Watch setup failed: %v\n", err)
			return
		}

		// 等待事件
		event := <-eventCh
		fmt.Printf("\nReceived event: %s\n", event.Type)

		if event.Type == zk.EventNodeDataChanged {
			data, _, err := conn.Get(path)
			if err != nil {
				fmt.Printf("Failed to get changed data: %v", err)
			} else {
				fmt.Printf("Updated data: %s\n", string(data))
			}
		}

		// 如果节点被删除则退出监听
		if event.Type == zk.EventNodeDeleted {
			fmt.Println("Node deleted. Stopping watch.")
			return
		}
	}
}
