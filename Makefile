deps:
	@echo "=== Installing dependencies ==="
	@dep ensure --vendor-only
	@echo "Done"

mocks:
	@echo "=== Generating Mocks ==="
	@rm -rf tests/mocks
	@CGO_ENABLED=0 ${GOPATH}/bin/mockery -all -dir domain/services -outpkg mocks -output tests/mocks/domain/services
	@echo "Hack because of vendoring googlemaps"
	sed -i "s/\github.com\/romeufcrosa\/where-to-eat\/vendor\///" tests/mocks/domain/services/GeoLocator.go

test: deps mocks
	@echo "=== Running tests ==="
	go test -cover ./...

.PHONY: coverage
coverage:
	go test -v -race -covermode=atomic -coverpkg=./... -coverprofile=coverage.txt ./...

build: deps
	@echo "=== Build image ==="
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o bin/api cmd/api/main.go

local-env:
	@echo "=== Starting local services ==="
	docker-compose up

local-run:
	@echo "=== Running application ==="
	go run cmd/api/main.go
