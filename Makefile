tgbot:
	go run ./cmd/tgbot

chat:
	cd ./cmd/chat && wails build

test:
	go test ./...

install:
	go mod tidy

check:
	go fmt ./... && go vet ./... && go test ./...

init_server: # create directories and configuration files on the service
	ssh $(host) "\
		mkdir -p /app/storage && \
		mkdir -p /var/log/files.md && \
		mkdir -p /opt/files.md && \
		chown -R www-data:www-data /app && \
		chown -R www-data:www-data /var/log/files.md && \
		chown -R www-data:www-data /opt/files.md && \
		echo 'BOT_API_TOKEN=' > /app/.env && \
		echo 'STORAGE_DIR=/app/storage' >> /app/.env && \
		echo 'SERVER_CERT_DIR=/opt/files.md' >> /app/.env && \
		echo 'SERVER_LOG_FILE=/var/log/files.md/server.log' >> /app/.env && \
		chown www-data:www-data /app/.env && \
		( \
			echo '[Unit]' > /etc/systemd/system/bot.service && \
			echo 'Description=Bot Service' >> /etc/systemd/system/bot.service && \
			echo 'After=network.target' >> /etc/systemd/system/bot.service && \
			echo '' >> /etc/systemd/system/bot.service && \
			echo '[Service]' >> /etc/systemd/system/bot.service && \
			echo 'User=www-data' >> /etc/systemd/system/bot.service && \
			echo 'ExecStart=/app/bot' >> /etc/systemd/system/bot.service && \
			echo 'WorkingDirectory=/app' >> /etc/systemd/system/bot.service && \
			echo 'Restart=always' >> /etc/systemd/system/bot.service && \
			echo 'RestartSec=5' >> /etc/systemd/system/bot.service && \
			echo 'StandardOutput=append:/app/log' >> /etc/systemd/system/bot.service && \
			echo 'StandardError=append:/app/err' >> /etc/systemd/system/bot.service && \
			echo 'AmbientCapabilities=CAP_NET_BIND_SERVICE' >> /etc/systemd/system/bot.service && \
			echo '' >> /etc/systemd/system/bot.service && \
			echo '[Install]' >> /etc/systemd/system/bot.service && \
			echo 'WantedBy=multi-user.target' >> /etc/systemd/system/bot.service \
		) || echo 'Failed to write service file. Check permissions.'; \
		echo 'Directories created and permissions set successfully.' \
	"

deploy: # deploy as systemd service
	@GREEN='\e[32m'; \
	YELLOW='\e[33m'; \
	RESET='\e[0m'; \
	printf "$${YELLOW}Building...$${RESET}\n" && \
	make check && \
	GOOS=linux GOARCH=amd64 go build -o /tmp/bot ./cmd/tgbot && \
	printf "$${GREEN}Build Completed$${RESET}\n" && \
	scp /tmp/bot $(host):/app/bot.new && printf "$${GREEN}The binary is copied on the server$${RESET}\n" && \
	ssh $(host) "mv /app/bot.new /app/bot && systemctl restart bot.service" && \
	rm /tmp/bot && \
	scp -r web files:/app && \
	printf "$${GREEN}Successfully deployed!$${RESET}\n"

deploy_binary: # deploy as regular binary
	@GREEN='\e[32m'; \
	YELLOW='\e[33m'; \
	RESET='\e[0m'; \
	printf "$${YELLOW}Building...$${RESET}\n" && \
	make check && \
	GOOS=linux GOARCH=amd64 go build -o /tmp/bot ./cmd/tgbot && \
	printf "$${GREEN}Build Completed$${RESET}\n" && \
	ssh $(host) "killall bot || true" && \
	scp /tmp/bot $(host):/app/bot && printf "$${GREEN}The binary is copied on the server$${RESET}\n" && \
	ssh $(host) "sudo setcap 'cap_net_bind_service=+ep' /app/bot" && \
	ssh $(host) "su -c \"cd /app && nohup ./bot >> /app/log 2>>/app/err &\" -s /bin/sh www-data" && \
	rm /tmp/bot && \
	printf "$${GREEN}Successfully deployed!$${RESET}\n"

lint:
	golangci-lint run

format:
	gofumpt -w .