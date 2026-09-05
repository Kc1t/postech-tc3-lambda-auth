FUNCTIONS := issuer authorizer

.PHONY: tidy test build package clean

tidy:
	go mod tidy

test:
	go test ./... -race -cover

build:
	@for fn in $(FUNCTIONS); do \
		GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o bin/$$fn/bootstrap ./cmd/$$fn; \
	done

package: build
	@for fn in $(FUNCTIONS); do \
		(cd bin/$$fn && zip -q ../../$$fn.zip bootstrap); \
	done

clean:
	rm -rf bin issuer.zip authorizer.zip
