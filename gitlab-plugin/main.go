package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"net/url"
)

// 定义项目响应体结构
type ProjectResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	WebURL      string `json:"web_url"`
	PathWithNamespace string `json:"path_with_namespace"`
	Namespace   struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Path string `json:"path"`
	} `json:"namespace"`
	CreatedAt   string `json:"created_at"`
	LastActivityAt string `json:"last_activity_at"`
}

// 定义项目列表响应结构
type ProjectsListResponse []ProjectResponse

func main() {
	// GitLab API URL
	gitlabURL := "https://gitlab-ce.alauda.cn/api/v4"

	// GitLab API Token - 从环境变量获取，避免硬编码
	token := os.Getenv("GITLAB_TOKEN")
	if token == "" {
		fmt.Println("错误: 请设置 GITLAB_TOKEN 环境变量")
		fmt.Println("使用方法: export GITLAB_TOKEN=your_token_here")
		os.Exit(1)
	}

	// 项目名称参数
	projectName := "hxia-test"

	// 构建获取项目的 URL (使用项目名称搜索)
	projectSearchURL := fmt.Sprintf("%s/projects", gitlabURL)

	// 创建查询参数
	params := url.Values{}
	params.Add("search", projectName)
	params.Add("simple", "false") // 获取详细信息
	params.Add("per_page", "10")  // 限制返回数量

	// 组合完整的 URL
	fullURL := fmt.Sprintf("%s?%s", projectSearchURL, params.Encode())

	// 创建 HTTP 请求
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		fmt.Printf("创建请求错误: %v\n", err)
		os.Exit(1)
	}

	// 设置请求头
	req.Header.Set("PRIVATE-TOKEN", token)
	req.Header.Set("Accept", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("发送请求错误: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应错误: %v\n", err)
		os.Exit(1)
	}

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("获取项目失败: %s\n状态码: %d\n响应体: %s\n", resp.Status, resp.StatusCode, string(body))
		os.Exit(1)
	}

	// 解析响应
	var projects ProjectsListResponse
	if err := json.Unmarshal(body, &projects); err != nil {
		fmt.Printf("解析响应错误: %v\n", err)
		os.Exit(1)
	}

	// 检查是否找到项目
	if len(projects) == 0 {
		fmt.Printf("未找到名称包含 '%s' 的项目\n", projectName)
		os.Exit(1)
	}

	// 输出项目信息
	fmt.Printf("找到 %d 个匹配的项目:\n", len(projects))
	fmt.Println("========================================")

	// 遍历输出所有匹配的项目
	for i, project := range projects {
		fmt.Printf("项目 %d:\n", i+1)
		fmt.Printf("ID: %d\n", project.ID)
		fmt.Printf("名称: %s\n", project.Name)
		fmt.Printf("描述: %s\n", project.Description)
		fmt.Printf("命名空间: %s (%s)\n", project.Namespace.Name, project.Namespace.Path)
		fmt.Printf("完整路径: %s\n", project.PathWithNamespace)
		fmt.Printf("项目URL: %s\n", project.WebURL)
		fmt.Printf("创建时间: %s\n", project.CreatedAt)
		fmt.Printf("最后活动时间: %s\n", project.LastActivityAt)
		fmt.Println("----------------------------------------")
	}

	// 尝试通过 ID 或路径获取特定项目的更多详细信息
	if len(projects) > 0 {
		firstProject := projects[0]
		// 可以选择通过项目 ID 或路径获取更详细的信息
		// 这里使用项目 ID 作为示例
		specificProjectURL := fmt.Sprintf("%s/projects/%d", gitlabURL, firstProject.ID)
		
		// 创建获取特定项目的请求
		specificReq, err := http.NewRequest("GET", specificProjectURL, nil)
		if err != nil {
			fmt.Printf("创建特定项目请求错误: %v\n", err)
			return
		}
		
		specificReq.Header.Set("PRIVATE-TOKEN", token)
		specificReq.Header.Set("Accept", "application/json")
		
		// 发送请求
		specificResp, err := client.Do(specificReq)
		if err != nil {
			fmt.Printf("发送特定项目请求错误: %v\n", err)
			return
		}
		defer specificResp.Body.Close()
		
		// 读取响应
		specificBody, err := ioutil.ReadAll(specificResp.Body)
		if err != nil {
			fmt.Printf("读取特定项目响应错误: %v\n", err)
			return
		}
		
		// 检查状态码
		if specificResp.StatusCode != http.StatusOK {
			fmt.Printf("获取特定项目失败: %s\n状态码: %d\n", specificResp.Status, specificResp.StatusCode)
			return
		}
		
		// 将响应保存到文件
		if err := ioutil.WriteFile("gitlab_project_response.json", specificBody, 0644); err != nil {
			fmt.Printf("保存响应到文件失败: %v\n", err)
		} else {
			fmt.Println("\n特定项目的完整响应已保存到 gitlab_project_response.json")
		}
	}
}