# v2c - An image analysis and provisioning workflow

## Documentation

* [Building components - enhancing the tool](docs/BUILDING-COMPONENTS.md)
* [Program architecture and how components work together](docs/DESIGN-AND-INTERFACES.md)

## Demo

This codebase ships with a demo of the proof of concept. The goal of the proof of contept was to demonstrate a workflow that shares the contents of a VMDK with a set of detective components which contribute material to a set of referenced image provisioners. Those provisioners transform the detective contributed materials and contribute Dockerfile segments. All of these contributions tar streamed via tar. The workflow finally stiches together contributed Dockerfile segments into a single Dockerfile and persists an expanded build context. The proof of concept does not perform final image assembly or actually use a real input image, detectives, or provisioners. Those components are crafted to demonstrate material contribution flow.

Start off by downloading this file:

    https://s3-us-west-2.amazonaws.com/allingeek-public-transport/for-export-flat.vmdk

Then clone this repo
    
    git clone https://github.com/docker/communitytools-image2docker-linux

Prepare the demo by running:

    make prepare
    make build
    make builtin-prep

Run on OSX with:

    sudo bin/v2c-darwin64 build -n PATH-TO_for-export-flat.vmdk

Run the demo on Linux with:

    sudo bin/v2c-linux64 build -n PATH-TO_for-export-flat.vmdk
