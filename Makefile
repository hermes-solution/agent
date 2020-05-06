.PHONY: clean build

VERSION=1.0.0

clean:
	rm -rf agent
	rm -rf agent.upx

build:
	env CGO_ENABLED=0 go build -ldflags="-s -w" . && upx --brute agent

docker: clean build
	docker build --no-cache  -t xuanloc0511/hermes-agent:${VERSION} .