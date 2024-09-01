run:
	go run ./cmd

test:
	go test ./...

install:
	go get ./...

check:
	go fmt ./... && go vet ./... && go test ./...

deploy:
	@GREEN='\e[32m'; \
    YELLOW='\e[33m'; \
    RESET='\e[0m'; \
    printf "$${YELLOW}Building...$${RESET}\n" && \
    make check && \
    GOOS=linux GOARCH=amd64 go build -o bot ./cmd && \
    printf "$${GREEN}Build Completed$${RESET}\n" && \
    ssh $(host) "killall bot || true" && \
    scp bot $(host):/app/bot && printf "$${GREEN}The binary is copied on the server$${RESET}\n" && \
    rm bot && \
    ssh $(host) "sudo setcap 'cap_net_bind_service=+ep' /app/bot" && \
    ssh $(host) "su -c \"cd /app && nohup ./bot >> /app/log 2>&1 &\" -s /bin/sh www-data" && \
    printf "$${GREEN}Successfully deployed!$${RESET}\n"