BINARY := mapitall
PREFIX ?= /usr/local

.PHONY: build vet test check install install-service cross clean

build:
	go build -o $(BINARY) ./cmd/mapitall

vet:
	go vet ./...

test:
	go test ./... -count=1

check: build vet test

install: build
	install -Dm755 $(BINARY) $(PREFIX)/bin/$(BINARY)

install-service: install
	@if [ "$$(uname)" = "Linux" ]; then \
		install -Dm644 mapitall.service /etc/systemd/user/mapitall.service; \
		install -Dm644 50-mapitall.rules /etc/udev/rules.d/50-mapitall.rules; \
		echo "Reload: systemctl --user daemon-reload && sudo udevadm control --reload"; \
	elif [ "$$(uname)" = "Darwin" ]; then \
		install -Dm644 com.hairglasses.mapitall.plist ~/Library/LaunchAgents/com.hairglasses.mapitall.plist; \
		echo "Load: launchctl load ~/Library/LaunchAgents/com.hairglasses.mapitall.plist"; \
	fi

cross:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY)-linux-amd64 ./cmd/mapitall
	GOOS=darwin GOARCH=arm64 go build -o $(BINARY)-darwin-arm64 ./cmd/mapitall
	GOOS=windows GOARCH=amd64 go build -o $(BINARY)-windows-amd64.exe ./cmd/mapitall

clean:
	rm -f $(BINARY) $(BINARY)-linux-* $(BINARY)-darwin-* $(BINARY)-windows-*
