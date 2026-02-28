GO_CMD=go
TEST_CMD=grc $(GO_CMD) test
COVERAGE_FILE=coverage.out

.PHONY: all test test-pkg test-cover clean

all: test

test:
	$(TEST_CMD) -v work

test-pkg:
	@if [ -z "$(PKG)" ]; then \
		echo "ERROR: Provide path to packet! Exampe: make test-pkg PKG=./libs/sx126x"; \
		exit 1; \
	fi
	$(TEST_CMD) -v $(PKG)

test-cover:
	$(TEST_CMD) -v -coverprofile=$(COVERAGE_FILE) work
	$(GO_CMD) tool cover -html=$(COVERAGE_FILE)

clean-cache:
	$(GO_CMD) clean -testcache
	@echo "Test cache cleaned"