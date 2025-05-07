VERSION=v1.0.0
build:
	@cd cmd && go mod tidy && CGO_ENABLED=0 go build -ldflags="-s -w" -o ../bin/memmq.bin

build-linux:
	@cd cmd && go mod tidy && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o ../bin/memmq.bin

docker-build:
	@docker build -t hellobchain/memmq:${VERSION} -f ./Dockerfile .

docker-build-linux:
	@cd cmd && go mod tidy && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o ../bin/memmq.bin
	@docker build -t dm/rwa-api:${VERSION} -f ./Dockerfile-linux .
