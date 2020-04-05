.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o components .
	upx components

.PHONY: docker
docker: build
	docker build -t ankrnetwork/opennet:latest .
	docker push ankrnetwork/opennet:latest