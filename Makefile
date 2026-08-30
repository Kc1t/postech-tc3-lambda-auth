BINARY := bootstrap
ARTIFACT := function.zip

.PHONY: tidy test build package clean

tidy:
	go mod tidy

test:
	go test ./... -race -cover

build:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o bin/$(BINARY) ./cmd/lambda

package: build
	cd bin && zip -q ../$(ARTIFACT) $(BINARY)

clean:
	rm -rf bin $(ARTIFACT)
