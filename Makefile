lint:
	@scripts/check_license.sh
	@scripts/run_golangci.sh

fmt:
	@scripts/run_gofmt.sh

include Makefile.common.mk


# Coverage tests
coverage:
	scripts/codecov.sh

vfsgen:
	go generate ./cmd/main.go
