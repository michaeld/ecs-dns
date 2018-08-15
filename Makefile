# Borrowed from: 
# https://github.com/silven/go-example/blob/master/Makefile
# https://vic.demuzere.be/articles/golang-makefile-crosscompile/

BINARY = ecs-dns
GOARCH = amd64

VERSION?=?
COMMIT=$(shell git rev-parse HEAD)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)

# Symlink into GOPATH
GITHUB_USERNAME=michaeld
BUILD_DIR=${GOPATH}/src/github.com/${GITHUB_USERNAME}/ecs-dns
CURRENT_DIR=$(shell pwd)
BUILD_DIR_LINK=$(shell readlink ${BUILD_DIR})

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS = -ldflags "-X github.com/michaeld/ecs-dns/cmd.GitSHA=${COMMIT}"

# Build the project
all: clean linux darwin

linux: 
	GOOS=linux GOARCH=${GOARCH} go build ${LDFLAGS} -o ./bin/${BINARY}-linux-${GOARCH} . ;

darwin:
	GOOS=darwin GOARCH=${GOARCH} go build ${LDFLAGS} -o ./bin/${BINARY}-darwin-${GOARCH} . ;

distro: linux docker

docker:
	docker build -t michaeld:ecs-dns-${VERSION} .   

fmt:
	go fmt $$(go list ./... | grep -v /vendor/) ; \

clean:
	-rm -rf ./bin

.PHONY: linux darwin fmt clean