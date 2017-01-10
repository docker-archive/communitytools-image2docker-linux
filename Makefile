PWD := $(shell pwd)

clean:
	@rm -f bin/v2c
	@-docker rmi docker/v2c:poc docker/v2c:latest docker/v2c:build-tooling

# Prepare tooling
prepare:
	@docker build -t docker/v2c:build-tooling -f tooling.df .

update-deps:
	@docker run --rm -v $(PWD):/go/src/github.com/docker/v2c -w /go/src/github.com/docker/v2c docker/v2c:build-tooling trash -u

update-vendor:
	@docker run --rm -v $(PWD):/go/src/github.com/docker/v2c -w /go/src/github.com/docker/v2c docker/v2c:build-tooling trash

iterate:
	@docker-compose -f iterate.dc kill
	@docker-compose -f iterate.dc rm
	@docker-compose -f iterate.dc up -d

fmt:
	# Formatting
	@docker run --rm \
	  -v $(PWD):/go/src/github.com/docker/v2c \
	  -w /go/src/github.com/docker/v2c \
	  docker/v2c:build-tooling \
	  go fmt

lint:
	# Linting
	@docker run --rm \
	  -v $(PWD):/go/src/github.com/docker/v2c \
	  -w /go/src/github.com/docker/v2c \
	  docker/v2c:build-tooling \
	  golint -set_exit_status

test:
	# Unit testing
	@docker run --rm \
	  -v $(PWD):/go/src/github.com/docker/v2c \
	  -v $(PWD)/bin:/go/bin \
	  -w /go/src/github.com/docker/v2c \
	  golang:1.7 \
	  go test
	# Test coverage
	@docker run --rm \
	  -v $(PWD):/go/src/github.com/docker/v2c \
	  -v $(PWD)/bin:/go/bin \
	  -w /go/src/github.com/docker/v2c \
	  golang:1.7 \
	  go test -cover

build: fmt lint test
	# Building binaries
	@docker run --rm \
	  -v $(PWD):/go/src/github.com/docker/v2c \
	  -v $(PWD)/bin:/go/bin \
	  -w /go/src/github.com/docker/v2c \
	  -e GOOS=linux \
	  -e GOARCH=amd64 \
	  golang:1.7 \
	  go build -o bin/v2c-linux64
	@docker run --rm \
	  -v $(PWD):/go/src/github.com/docker/v2c \
	  -v $(PWD)/bin:/go/bin \
	  -w /go/src/github.com/docker/v2c \
	  -e GOOS=darwin \
	  -e GOARCH=amd64 \
	  golang:1.7 \
	  go build -o bin/v2c-darwin64

release: build
	@docker build -t docker/v2c:latest -f release.df .
	@docker tag docker/v2c:latest docker/v2c:poc

democlean:
	@docker rmi v2c/packager:demo v2c/os.ubuntu-detective:v16.04 v2c/app.random-detective:1 v2c/ubuntu-provisioner:v16.04 v2c/app.random.provisioner:1

demoprep: democlean
	@docker build -t v2c/packager:demo -f ./packager/Packager.df ./packager/
	@docker build -t v2c/os.ubuntu-detective:v16.04 -f ./detectives/os.ubuntu16.04.df ./detectives/
	@docker build -t v2c/app.random-detective:1 -f ./detectives/app.random1.df ./detectives/
	@docker build -t v2c/ubuntu-provisioner:v16.04 -f ./provisioners/os.ubuntu16.04.df ./provisioners/
	@docker build -t v2c/app.random.provisioner:1 -f ./provisioners/app.random1.df ./provisioners/

demo-darwin: build 
	@bin/v2c-darwin64 build demo.vmdk
demo-linux: build 
	@bin/v2c-linux64 build demo.vmdk