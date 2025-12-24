package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

func main() {
	// 设置 Elasticsearch 客户端配置
	cfg := elasticsearch.Config{
		Addresses: []string{
			"https://192.168.138.215/es_proxy", // 默认的 Elasticsearch 地址
		},
		// 如果需要认证，可以取消下面的注释并修改用户名和密码
		Username: "RnpJKiLd",
		Password: "tyjWoJfs8vcg1FAe27lZ6dBu761lz5IA",
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// 创建客户端
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("创建 Elasticsearch 客户端失败: %s", err)
	}

	// 检查集群是否可用
	res, err := es.Info()
	if err != nil {
		log.Fatalf("获取集群信息失败: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Fatalf("获取集群信息返回错误: %s", res.Status())
	}

	// 定义 ILM 策略的名称
	policyName := "my-ilm-policy"

	// 定义 ILM 策略的 JSON 内容
	policyBody := `{
  "policy": {
    "phases": {
      "hot": {
        "actions": {
          "rollover": {
            "max_size": "50gb",
            "max_age": "30d"
          }
        }
      },
      "delete": {
        "min_age": "90d",
        "actions": {
          "delete": {}
        }
      }
    }
  }
}`

	// 创建 ILM 策略
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	createReq := esapi.ILMPutLifecycleRequest{
		Policy: policyName,
		Body:   strings.NewReader(policyBody),
	}

	createRes, err := createReq.Do(ctx, es)
	if err != nil {
		log.Fatalf("创建 ILM 策略失败: %s", err)
	}
	defer createRes.Body.Close()

	if createRes.IsError() {
		log.Fatalf("创建 ILM 策略返回错误: %s", createRes.Status())
	}

	fmt.Printf("成功创建 ILM 策略 '%s'\n", policyName)

	// 验证策略是否创建成功
	getReq := esapi.ILMGetLifecycleRequest{
		Policy: policyName, // 根据错误提示，Policy参数应该是字符串类型而不是[]string类型
	}

	getRes, err := getReq.Do(ctx, es)
	if err != nil {
		log.Fatalf("获取 ILM 策略失败: %s", err)
	}
	defer getRes.Body.Close()

	if getRes.IsError() {
		log.Fatalf("获取 ILM 策略返回错误: %s", getRes.Status())
	}

	// 创建索引模板，将ILM策略应用于特定索引
	// 索引模板名称
	templateName := "my-index-template"

	// 关于索引别名在索引模板中的创建：
	// 1. Elasticsearch 7.9及以上版本支持在索引模板中直接定义索引别名
	// 2. 在模板中使用"aliases"字段可以为匹配的索引自动创建别名
	// 3. "index.lifecycle.rollover_alias"是ILM滚动操作专用的配置，不创建实际别名

	// 定义索引模板的JSON内容（包含别名定义）
	templateBody := `{
  "index_patterns": ["logstash-*"], // 只匹配以logstash-开头的索引
  "template": {
    "settings": {
      "index.lifecycle.name": "` + policyName + `", // 应用我们创建的ILM策略
      "index.lifecycle.rollover_alias": "logstash" // ILM滚动操作的别名
    },
    "aliases": {
      "all-logs": {}, // 为匹配的索引创建一个名为all-logs的别名
      "logstash-write": { // 写入别名
        "is_write_index": false // 不设置为写入索引（ILM滚动时自动管理）
      }
    }
  }
}`

	// 创建索引模板请求
	templateReq := esapi.IndicesPutTemplateRequest{
		Name: templateName,
		Body: strings.NewReader(templateBody),
	}

	templateRes, err := templateReq.Do(ctx, es)
	if err != nil {
		log.Fatalf("创建索引模板失败: %s", err)
	}
	defer templateRes.Body.Close()

	if templateRes.IsError() {
		log.Fatalf("创建索引模板返回错误: %s", templateRes.Status())
	}

	fmt.Printf("成功创建索引模板 '%s'，该模板将ILM策略应用于匹配'logstash-*'模式的索引\n", templateName)
	fmt.Printf("索引模板中已定义以下别名：\n")
	fmt.Printf("- all-logs：用于读取所有匹配模板的日志索引\n")
	fmt.Printf("- logstash-write：可用于写入操作（配合rollover_alias使用）\n")
	fmt.Printf("注意：对于ILM的rollover功能，还需要手动创建初始索引并关联别名\n")

	// 另一种方法：直接为现有索引应用ILM策略
	// 如果您想为特定的现有索引应用ILM策略，可以使用以下代码
	// 注意：取消下面的注释并修改indexName为您要应用策略的索引名称
	/*
		indexName := "your-specific-index"
		updateIndexReq := esapi.IndicesPutSettingsRequest{
			Index: []string{indexName},
			Body:  strings.NewReader(fmt.Sprintf(`{"index.lifecycle.name": "%s"}`, policyName)),
		}

		updateIndexRes, err := updateIndexReq.Do(ctx, es)
		if err != nil {
			log.Fatalf("更新索引设置失败: %s", err)
		}
		defer updateIndexRes.Body.Close()

		if updateIndexRes.IsError() {
			log.Fatalf("更新索引设置返回错误: %s", updateIndexRes.Status())
		}

		fmt.Printf("成功为索引 '%s' 应用ILM策略 '%s'\n", indexName, policyName)
	*/

	// 演示：通过索引别名查找对应的所有索引名
	fmt.Println("\n演示: 通过索引别名查找对应的所有索引名")

	// 1. 获取所有索引别名及其关联的索引
	allAliases, err := getIndicesByAlias(ctx, es, "*")
	if err != nil {
		log.Printf("获取所有别名信息时出错: %s", err)
	} else {
		fmt.Println("当前系统中的所有别名及其关联索引:")
		for alias, indices := range allAliases {
			fmt.Printf("别名 '%s' 关联的索引: %s\n", alias, strings.Join(indices, ", "))
		}
	}

	// 2. 获取特定别名关联的索引
	specificAlias := "all-logs"
	specificAliases, err := getIndicesByAlias(ctx, es, specificAlias)
	if err != nil {
		log.Printf("获取特定别名 '%s' 信息时出错: %s", specificAlias, err)
	} else if indices, exists := specificAliases[specificAlias]; exists {
		fmt.Printf("\n别名 '%s' 关联的索引: %s\n", specificAlias, strings.Join(indices, ", "))
	} else {
		fmt.Printf("\n别名 '%s' 不存在或未关联任何索引\n", specificAlias)
	}

	fmt.Printf("\n成功获取 ILM 策略 '%s'，策略已正确创建并通过索引模板应用于特定索引\n", policyName)
}

// getIndicesByAlias 获取指定别名或所有别名关联的索引信息
// 参数:
//
//	ctx: 上下文，用于控制请求超时
//	es: Elasticsearch客户端实例
//	aliasName: 要查询的别名名称，使用"*"可以查询所有别名
//
// 返回值:
//
//	map[string][]string: 键为别名名称，值为该别名关联的索引名列表
//	error: 操作过程中的错误信息
func getIndicesByAlias(ctx context.Context, es *elasticsearch.Client, aliasName string) (map[string][]string, error) {
	// 创建获取别名的请求
	req := esapi.IndicesGetAliasRequest{
		Name: []string{aliasName}, // 如果为空或使用"*"，则获取所有别名
	}

	// 执行请求
	res, err := req.Do(ctx, es)
	if err != nil {
		return nil, fmt.Errorf("获取别名信息失败: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("获取别名信息返回错误: %s", res.Status())
	}

	// 解析响应
	var result map[string]map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析别名信息失败: %w", err)
	}

	// 构建别名到索引的映射
	aliasToIndices := make(map[string][]string)
	for indexName, data := range result {
		if aliases, ok := data["aliases"].(map[string]interface{}); ok {
			for alias := range aliases {
				aliasToIndices[alias] = append(aliasToIndices[alias], indexName)
			}
		}
	}

	return aliasToIndices, nil
}
