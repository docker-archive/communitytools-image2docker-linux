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
	@docker rmi $(docker images --filter label=com.docker.v2c.component.demo -aq)

demoprep:
	@docker build -t v2c/packager:demo -f ./packager/Packager.df ./packager/
	@docker build -t v2c/app.random-detective:1 -f ./detectives/app.random1.df ./detectives/
	@docker build -t v2c/app.random-detective:2 -f ./detectives/app.random2.df ./detectives/
	@docker build -t v2c/app.random-detective:3 -f ./detectives/app.random3.df ./detectives/
	@docker build -t v2c/app.random.provisioner:1 -f ./provisioners/app.random1.df ./provisioners/
	@docker build -t v2c/app.random.provisioner:2 -f ./provisioners/app.random2.df ./provisioners/
	@docker build -t v2c/app.random.provisioner:3 -f ./provisioners/app.random3.df ./provisioners/

demo-darwin:
	@bin/v2c-darwin64 build demo.vmdk
demo-linux:
	@bin/v2c-linux64 build demo.vmdk

builtins:
	@docker build -t v2c/centos-detective:v6.8 -f ./detectives/os.centos6.8.df ./detectives/
	@docker build -t v2c/centos-provisioner:v6.8 -f ./provisioners/os.centos6.8.df ./provisioners/
	@docker build -t v2c/ubuntu-detective:v16.04 -f ./detectives/os.ubuntu16.04.df ./detectives/
	@docker build -t v2c/ubuntu-provisioner:v16.04 -f ./provisioners/os.ubuntu16.04.df ./provisioners/
	@docker build -t v2c/ubuntu-detective:v14.04.5 -f ./detectives/os.ubuntu14.04.5.df ./detectives/
	@docker build -t v2c/ubuntu-provisioner:v14.04.5 -f ./provisioners/os.ubuntu14.04.5.df ./provisioners/
	@docker build -t v2c/app.apt-repl.detective:1 -f ./detectives/app.apt-repl-nover.df ./detectives/
	@docker build -t v2c/app.apt-repl.provisioner:1 -f ./provisioners/app.apt-repl.df ./provisioners/
