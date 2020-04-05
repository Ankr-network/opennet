.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o opennet .
	upx opennet

.PHONY: clean
clean:
	rm -rf opennet