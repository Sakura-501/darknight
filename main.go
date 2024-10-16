package main

import (
	"fmt"
)

type UploadRequest struct {
	Message string `json:"message"`
	Content string `json:"content"`
	SHA     string `json:"sha,omitempty"` // SHA is optional for new files
}

func upload_file(token, user, repo string) {

}

func main() {
	fmt.Println("Welcome to DarkNight")
	//token := "xxxx"
	//user := "Sakura-501"
	//repo := "github-c2-test"
}
