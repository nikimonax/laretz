BIN_DIR := ./bin

COV_FILE := coverage.out
COV_FILE_HTML := coverage.html

EXTRA_ARGS += $(ARGS)


all: build


.SUFFIXES:

.PHONY: .FORCE
.FORCE:

.PHONY: run
run: $(BIN_DIR)/laretz
	./$< $(EXTRA_ARGS)

.PHONY: gen generate
gen generate:
	sqlc generate
	templ generate
	mockery

.PHONY: build
build: $(BIN_DIR)/laretz

.PHONY: lint
lint: _check_golangci_lint_cmd
	golangci-lint run

.PHONY: fmt format
fmt format: _check_golangci_lint_cmd
	golangci-lint fmt

.PHONY: test
test:
	go test $(EXTRA_ARGS) $$(go list ./... | grep -vE "internal/testing|database/generated")

.PHONY: cover
cover: $(COV_FILE)
	sed -i '/internal\/testing/d' $(COV_FILE) && \
	go tool cover -func=$(COV_FILE)

.PHONY: cover-html
cover-html: $(COV_FILE)
	go tool cover -html=$(COV_FILE) -o $(COV_FILE_HTML)

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)
	go clean -testcache

$(COV_FILE): EXTRA_ARGS += -coverpkg=./... -coverprofile=$(COV_FILE)
$(COV_FILE): test

$(BIN_DIR)/laretz: $(BIN_DIR)/%: cmd/% .FORCE
	go build -o $@ ./$<


.PHONY: _check_golangci_lint_cmd
_check_golangci_lint_cmd:
	@if ! command -v golangci-lint $&> /dev/null; then \
		echo "[ERROR]: command 'golangci-lint' not found"; \
		echo "install using \"go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest\""; \
		exit 1; \
	fi
