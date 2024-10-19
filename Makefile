all: github-implant github-teamserver


github-implant: github/socks5/*.go github/implant/github-implant.go
	go build -o github-implant github/implant/github-implant.go

github-teamserver: github/socks5/*.go github/teamserver/github-teamserver.go
	go build -o github-teamserver github/teamserver/github-teamserver.go

