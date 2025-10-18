BINARY_NAME=shutdown-service
INSTALL_PATH=/usr/local/bin/$(BINARY_NAME)
SERVICE_NAME=shutdown-service
SERVICE_FILE=/etc/systemd/system/$(SERVICE_NAME).service
ENV_FILE=/etc/shutdown-service.env

.PHONY: build install deploy restart status logs

build:
	go build -o $(BINARY_NAME) .

install: build
	sudo cp $(BINARY_NAME) $(INSTALL_PATH)
	sudo chmod +x $(INSTALL_PATH)

service-file:
	@echo "Creating systemd service file..."
	@echo "[Unit]\nDescription=Home Server Shutdown Service\nAfter=network.target\n\n[Service]\nType=simple\nUser=$(USER)\nEnvironmentFile=$(ENV_FILE)\nExecStart=$(INSTALL_PATH)\nRestart=on-failure\nRestartSec=5s\n\n[Install]" | sudo tee $(SERVICE_FILE) > /dev/null
	@echo "WantedBy=multi-user.target" | sudo tee -a $(SERVICE_FILE) > /dev/null

deploy: install service-file
	sudo systemctl stop $(SERVICE_NAME) 2>/dev/null || true
	sudo systemctl daemon-reload
	sudo systemctl enable $(SERVICE_NAME)
	sudo systemctl restart $(SERVICE_NAME)

restart:
	sudo systemctl restart $(SERVICE_NAME)

status:
	sudo systemctl status $(SERVICE_NAME)

logs:
	sudo journalctl -u $(SERVICE_NAME) -f