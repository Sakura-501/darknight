package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/xen0n/go-workwx"
	"net/http"
	"os"
	"text/template"
)

func bot_send_text_message(webhook_url string, message_text string) error {
	var buf bytes.Buffer
	template.JSEscape(&buf, []byte(message_text))
	encodedString := buf.String()
	// 创建请求体的JSON数据
	requestBody := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": fmt.Sprintf("%s", encodedString),
		},
	}
	// 将请求体编码为JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("Error encoding JSON: %s\n", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", webhook_url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("Error creating request: %s\n", err)
	}

	// 设置请求头，指定内容类型为JSON
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error sending request: %s\n", err)
	}
	defer resp.Body.Close()
	//// 读取响应体
	//body, _ := ioutil.ReadAll(resp.Body)
	//// 解析响应体中的JSON数据
	//var responseBody string
	//err = json.Unmarshal(body, &responseBody)
	//fmt.Println(responseBody)

	return nil
}

func create_app_chat_in_static_chatid(user_id string) string {
	chat_id := "wechat-c2-test"
	var chatinfo workwx.ChatInfo
	chatinfo.ChatID = chat_id
	chatinfo.OwnerUserID = user_id
	chatinfo.MemberUserIDs = []string{user_id}
	chatID, err := app.CreateAppchat(&chatinfo)
	fmt.Println(chatID)
	fmt.Println(err)

	return chatID
}

var app *workwx.WorkwxApp

// 已寄
func main() {
	fmt.Println("Welcome to the DarkNight")
	if len(os.Args) < 2 {
		fmt.Println("Usage: implant <corp_id> <corp_secret> <user_id>")
		return
	}
	//webhook_key := os.Args[1]
	//webhook_url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=%s", webhook_key)
	corp_id := os.Args[1]
	corp_secret := os.Args[2]
	user_id := os.Args[3]
	agentID := int64(123456789)
	client := workwx.New(corp_id)
	app = client.WithApp(corp_secret, agentID)
	chat_id := create_app_chat_in_static_chatid(user_id)
	fmt.Println(chat_id)

}
