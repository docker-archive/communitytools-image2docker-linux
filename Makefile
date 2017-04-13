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
	  ./scripts/fmt.sh

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
	@docker rmi $(shell docker images --filter label=com.docker.v2c.component.demo -aq) &>/dev/null || true

demoprep:
	@docker build -t v2c/packager-demo:1 -f ./packager/demo/Packager.df ./packager/demo/
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

builtin-clean:
	@docker rmi $(shell docker images --filter label=com.docker.v2c.component.builtin -aq) &>/dev/null || true

builtin-prep:
	@docker build -t v2c/guestfish-export:1 -f ./packager/guestfish-export.df ./packager/

	@docker build -t v2c/centos-detective:v6.8   -f ./detectives/os.centos6.8.df ./detectives/
	@docker build -t v2c/centos-provisioner:v6.8 -f ./provisioners/os.centos6.8.df ./provisioners/

	@docker build -t v2c/ubuntu-detective:v16.04   -f ./detectives/os.ubuntu16.04.df ./detectives/
	@docker build -t v2c/ubuntu-provisioner:v16.04 -f ./provisioners/os.ubuntu16.04.df ./provisioners/

	@docker build -t v2c/ubuntu-detective:v16.10   -f ./detectives/os.ubuntu16.10.df ./detectives/
	@docker build -t v2c/ubuntu-provisioner:v16.10 -f ./provisioners/os.ubuntu16.10.df ./provisioners/

	@docker build -t v2c/ubuntu-detective:v14.04.5   -f ./detectives/os.ubuntu14.04.5.df ./detectives/
	@docker build -t v2c/ubuntu-provisioner:v14.04.5 -f ./provisioners/os.ubuntu14.04.5.df ./provisioners/

	@docker build -t v2c/app.apt-repl.detective:1   -f ./detectives/app.apt-repl-nover.df ./detectives/
	@docker build -t v2c/app.apt-repl.provisioner:1 -f ./provisioners/app.apt-repl.df ./provisioners/

	@docker build -t v2c/conf.apache2-var-www.detective:1   -f ./detectives/conf.apache2-var-www.df ./detectives/
	@docker build -t v2c/conf.apache2-var-www.provisioner:1 -f ./provisioners/conf.apache2-var-www.df ./provisioners/

	@docker build -t v2c/conf.apache2-etc.detective:1   -f ./detectives/conf.apache2-etc.df   ./detectives/
	@docker build -t v2c/conf.apache2-etc.provisioner:1 -f ./provisioners/conf.apache2-etc.df ./provisioners/

	@docker build -t v2c/conf.mysql5-data.detective:1   -f ./detectives/conf.mysql5-data.df ./detectives/
	@docker build -t v2c/conf.mysql5-data.provisioner:1 -f ./provisioners/conf.mysql5-data.df ./provisioners/

	@docker build -t v2c/runit-detective:ubuntu-v14.04.5    -f ./detectives/init.ubuntu14.04.5.df ./detectives/
	@docker build -t v2c/runit-provisioner:ubuntu-v14.04.5  -f ./provisioners/init.ubuntu14.04.5.df ./provisioners/

	@docker build -t v2c/runit-detective:ubuntu-v16.04    -f ./detectives/init.ubuntu16.04.df   ./detectives/
	@docker build -t v2c/runit-provisioner:ubuntu-v16.04  -f ./provisioners/init.ubuntu16.04.df ./provisioners/

	@docker build -t v2c/runit-detective:ubuntu-v16.10    -f ./detectives/init.ubuntu16.10.df   ./detectives/
	@docker build -t v2c/runit-provisioner:ubuntu-v16.10  -f ./provisioners/init.ubuntu16.10.df ./provisioners/

	@docker build -t v2c/init.apache2-sysv.detective:2 -f ./detectives/init.apache2-sysv.df ./detectives/
	@docker build -t v2c/init.apache2-provisioner:2    -f ./provisioners/init.apache2.df    ./provisioners/

	@docker build -t v2c/app.tomcat8.5.5-detective:1   -f ./detectives/app.tomcat8.5.5.df   ./detectives/
	@docker build -t v2c/app.tomcat8.5.5-provisioner:1 -f ./provisioners/app.tomcat8.5.5.df ./provisioners/
	@docker build -t v2c/init.tomcat-detective:1   -f ./detectives/init.tomcat-systemd.df   ./detectives/
	@docker build -t v2c/init.tomcat-provisioner:1 -f ./provisioners/init.tomcat-systemd.df ./provisioners/
