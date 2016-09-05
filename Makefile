version = 1.0
all: pacyak-$(version)-linux-amd64 pacyak-$(version)-darwin-amd64 pacyak-$(version)-windows-amd64

pacyak-$(version)-linux-amd64:
	GOARCH=amd64 GOOS=linux go build -o $@

pacyak-$(version)-darwin-amd64:
	GOARCH=amd64 GOOS=darwin go build -o $@

pacyak-$(version)-windows-amd64:
	GOARCH=amd64 GOOS=windows go build -o $@

