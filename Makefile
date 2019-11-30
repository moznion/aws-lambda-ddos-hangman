.PHONY: deps clean build

PKGS := $(shell go list ./...)

build:
	GOOS=linux GOARCH=amd64 go build -o execute/execute ./execute

clean:
	rm -rf ./execute/execute

deploy: build
	sam package \
		--s3-bucket $(S3BUCKET) \
		--output-template-file .packaged.yaml \
		--template-file "template.yaml"

check: test lint vet fmt-check

test:
	go test -v -race $(PKGS)

lint:
	golint -set_exit_status $(PKGS)

vet:
	go vet $(PKGS)

fmt-check:
	gofmt -l -s execute/*.go | grep [^*][.]go$$; \
	EXIT_CODE=$$?; \
	if [ $$EXIT_CODE -eq 0 ]; then exit 1; fi; \
	goimports -l execute/*.go | grep [^*][.]go$$; \
	EXIT_CODE=$$?; \
	if [ $$EXIT_CODE -eq 0 ]; then exit 1; fi \

