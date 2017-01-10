# v2c - An image analysis and provisioning workflow

## Demo

This codebase ships with a demo of the proof of concept. The goal of the proof of contept was to demonstrate a workflow that shares the contents of a VMDK with a set of detective components which contribute material to a set of referenced image provisioners. Those provisioners transform the detective contributed materials and contribute Dockerfile segments. All of these contributions tar streamed via tar. The workflow finally stiches together contributed Dockerfile segments into a single Dockerfile and persists an expanded build context. The proof of concept does not perform final image assembly or actually use a real input image, detectives, or provisioners. Those components are crafted to demonstrate material contribution flow.

Prepare the demo by running:

    make demoprep

Run the demo on OSX with:

    make demo-darwin

Run the demo on Linux with:

    make demo-linux
