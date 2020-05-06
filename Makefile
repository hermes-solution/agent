.PHONY: clean build

VERSION=1.0.0

clean:
	rm -rf ./docker_build/agent
	rm -rf ./docker_build/agent.upx

build:
	env CGO_ENABLED=0 go build -ldflags="-s -w" -o docker_build/agent . && upx --brute ./docker_build/agent
	rm -rf ./docker_build/agent.upx

docker: clean build
	docker build --no-cache  -t xuanloc0511/hermes-agent:${VERSION} ./docker_build