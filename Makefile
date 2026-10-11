# kopi-o: `make setup build test run package install`
BIN := bin/kopi-o
APP := bin/Kopi-O.app

.PHONY: setup build test run package install clean

setup: ## fetch module deps
	go mod tidy

build: ## local binary (needs CLT clang for systray cgo)
	mkdir -p bin
	CGO_ENABLED=1 go build -o $(BIN) ./cmd/kopi-o

test: ## unit tests (config, supervisor; no tray needed)
	CGO_ENABLED=1 go test ./...

run: build ## run in foreground
	./$(BIN) -foreground

package: build ## assemble Kopi-O.app bundle (LSUIElement) + ad-hoc sign + zip
	./scripts/package.sh

install: package ## install to ~/Applications + optional LaunchAgent
	./scripts/install.sh

clean:
	rm -rf bin
