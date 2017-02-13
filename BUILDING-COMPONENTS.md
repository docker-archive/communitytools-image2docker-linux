# Writing Detectives and Provisioners

This project's architecture and the relationship between components is covered in [DESIGN-AND-INTERFACES](DESIGN-AND-INTERFACES.md). This document walks the reader through building their own detectives and provisioners. If you're starting a lift-and-shift project of your own and need to capture something that current plugins miss then this document is for you.

## Detectives

The purpose of a detective is to inspect the source file system and capture any materials that a provisioner will need to realize some target component in the resulting Docker image. Examples of material that a detective might emit include:

* a list of package names for installation with yum or apt-get
* a tar archive with a set of configuration files
* a tar archive containing a custom web application

In all of these cases the detective shares this material by writing to its STDOUT stream and exiting with a status code of 0. If a detective exits with a 0 then the v2c workflow will start the related provisioner and send the material to the STDIN stream for that provisioner.

Detectives only determine if provisioning should occur and optionally pass along custom material from the input disk image. That is all. Installing software and placing custom materials into the resulting Docker image is a job for provisioners.

### A Simple Detective

The best detectives detect and capture something specific and do so with very minor tooling. Remember that v2c will run all detectives in parallel so feel free to create specialized detectives and leverage that parallelism to capture more nuanced cases. Don't feel the need to write one massive detective that handles every case.

Consider an operating system detective. Most Linux distributions can be identified with minimal effort. But different distributions use different mechanisms to provide that identification. So any detective we build for this purpose will need some minimal tooling for working with the files on the input image. Let's use ````alpine```` in this case.

    FROM alpine:latest

Suppose we're specifically targetting Ubuntu 14.04.5 with this detective. We know - through basic research - that Ubuntu uses a file at ````/etc/os-release```` to store release metadata and that the file contains a line prefixed (keyed) with "PRETTY_NAME" which is unique for each release (and also happens to be human readable). After a bit more research we've determined that the exact value following ````PRETTY_NAME```` for Ubuntu 14.04.5 is ````Ubuntu 14.04.5 LTS````. So our new detective needs to run a program that determines if this file on the input disk matches that pattern and return a 0 if it does.

The alpine image includes grep. Grep is a popular tool for listing files that contain string patterns. Grep also happens to exit with a 0 if there was a match and a 1 otherwise. These attributes make ````grep```` a perfect tool for our detective.

    FROM alpine:latest
    CMD grep "PRETTY_NAME=\"Ubuntu 14.04.5 LTS\"" /etc/os-release

Unfortunately, if you tried to build and run a detective from the Dockerfile above you'd find that there are a few problems. First, its not a detective yet. We'll address that in a moment. Second, and more importantly, if you had somehow gotten it to run you'd notice that the file ````/etc/os-release```` does not exist and that the grep command wrote a bunch of data to STDOUT or STDERR. A detective should only write data to STDOUT if it needs to share something with its provisioner. Before we actually make this a detective lets fix those bugs.

The reason grep cannot find /etc/os-release is that the input filesystem is mounted in at ````/v2c/disk````. That means that the file grep needs to inspect is actually at ````/v2c/disk/etc/os-release````.

Since we're matching against a specific file we don't really care about any of grep's output. To suppress it all we can use stream redirection and push it all to /dev/null.

    FROM alpine:latest
    CMD grep "PRETTY_NAME=\"Ubuntu 14.04.5 LTS\"" /v2c/disk/etc/os-release 1>2 2>/dev/null

With those bugs fixed we're ready to add the label metadata that v2c requires to identify detectives and know how to use them. There are four that should be added:

1. com.docker.v2c.component
2. com.docker.v2c.component.category
3. com.docker.v2c.component.rel
4. com.docker.v2c.component.description

The first label is used to identify the component type. Detectives should provide ````detective```` here. The second is the category of this detective. Categories are fixed and determine the order of contributions in the processing and provisioning workflow. The appropriate category for this detective is ````os````. The third label is ````rel```` and is an image identifier for the provisioner that should be executed if this detective returns a 0. The last label is an opporutnity to describe the purpose of the detective and what it targets.

Putting all this together in a Dockerfile will look like this:

    FROM alpine:latest
    LABEL com.docker.v2c.component=detective \
          com.docker.v2c.component.category=os \
          com.docker.v2c.component.rel=v2c/ubuntu-provisioner:v14.04.5 \
          com.docker.v2c.component.description=Detects\ the\ Trusty\ Tahr\ release\ of\ Ubuntu
    CMD grep "PRETTY_NAME=\"Ubuntu 14.04.5 LTS\"" /v2c/disk/etc/os-release 1>2 2>/dev/null

