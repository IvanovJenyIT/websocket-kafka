build-chat:
	@go build -o ./bin/chat ./...
	@chmod +x ./bin/chat

chat: build-chat
	@./bin/chat

text-chat-race:
	@go clean -testcache
	@go test -race -v ./...