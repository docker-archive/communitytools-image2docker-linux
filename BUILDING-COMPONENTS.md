# Writing Detectives and Provisioners

This project's architecture and the relationship between components is covered in [DESIGN-AND-INTERFACES](DESIGN-AND-INTERFACES.md). This document walks the reader through building their own detectives and provisioners. If you're starting a lift-and-shift project of your own and need to capture something that current plugins miss then this document is for you.

## Detectives

The purpose of a detective is to inspect the source file system and capture any materials that a provisioner will need to realize some target component in the resulting Docker image. Examples of material that a detective might emit include:

* a list of package names for installation with yum or apt-get
* a tar archive with a set of configuration files
* a tar archive containing a custom web application

In all of these cases the detective shares this material by writing to its STDOUT stream and exiting with a status code of 0. If a detective exits with a 0 then the v2c workflow will start the related provisioner and send the material to the STDIN stream for that provisioner.

Detectives only determine if provisioning should occur and optionally pass along custom material from the input disk image. That is all. Installing software and placing custom materials into the resulting Docker image is a job for provisioners.

## A Simple Detective

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

## A Simple Provisioner

The simple OS detective in the previous section really only signals that a specific provisioner should be executed. No materials are provided from the original disk image. In some cases like selecting a base image none are required. Provisioners in the ````os```` category are particularly simple. Like detectives a good provisioner should make very specific contributions to the resulting Dockerfile and filesystem. In this case the only contribution that needs to be made is a ````FROM```` instruction to the Dockerfile.

All provisioners provide results to the workflow by writing a tar archive to STDOUT. Contributions to the final Dockerfile are communicated by including them in a Dockerfile at the "root" of that tar archive. An example might ease the confusion you're likely experiencing right now.

Create a file named ````./os.ubuntu14.04.5/Dockerfile```` and drop in the instruction that you want to provide for final image assembly:

    FROM ubuntu:14.04.5
    ...

Now create a Dockerfile for your provisioner named ````./os.ubuntu14.04.5.df```` and start putting together the provisioner. Again, the provisioner will need some basic system tools like a shell, and tar so ````alpine```` is a great base image to start with. Since this provisioner will always contribute the same Dockerfile fragment you could either generate that fragment on the fly each time before generating the tar archive or just add it to the image and generate the tar archive at build time. In this case just add the fragment to the root of the provisioner image at ````/Dockerfile```` and create the tar during image build.

    FROM alpine
    ...
    COPY ./os.ubuntu14.04.5/Dockerfile /Dockerfile
    RUN tar cf payload.tar Dockerfile
    ...

Now that the output material has been included with the provisioner you need to determine how to send that content to STDOUT at runtime. The ````cat```` command is a great way to accomplish this task.

    FROM alpine
    ...
    COPY ./os.ubuntu14.04.5/Dockerfile /Dockerfile
    RUN tar cf payload.tar Dockerfile
    ENTRYPOINT ["/bin/sh"]
    CMD ["-c", "cat /payload.tar"]

Before v2c will discover and use the new provisioner you need to add the label metadata. Provisioners specify ````provisioner```` for the component label and the category should match that provided by the detective.

    FROM alpine
    LABEL com.docker.v2c.component=provisioner \
          com.docker.v2c.component.category=os \
          com.docker.v2c.component.description=Provisions\ Ubuntu\ Trusty\ Tahr\ images
    COPY ./os.ubuntu14.04.5/Dockerfile /Dockerfile
    RUN tar cf payload.tar Dockerfile
    ENTRYPOINT ["/bin/sh"]
    CMD ["-c", "cat /payload.tar"]

## Starting Services Inside the Container

Replicating the behavior you'd expect when a virtual machine boots is tricky in a container. Containers are designed to isolate single processes or a collection of processes. As such Docker and containers are payload agnostic. If you wish to start a collection of services when you launch a container, that container needs to bring its own init system. That system should be both capable of service monitoring and proper signal handling.

The tight coupling between the resulting Dockerfile, the init system, and generalized component detection makes building a general solution particularly challenging. This lift-and-shift project takes an opinionated approach. This project creates images that use runit for an init system and runs detectives and provisioners in the ````init```` category.

This opinion is flexible. Since there are many methods of installing ````runit```` onto a target operating system image we require that this init system is provisioned similar to operating systems. For example...

#### Ubuntu 14.04.5 Init Detective

    FROM alpine:3.4
    LABEL com.docker.v2c.component=detective \
          com.docker.v2c.component.category=init \
          com.docker.v2c.component.builtin=1 \
          com.docker.v2c.component.rel=v2c/runit-provisioner:ubuntu-v14.04.5 \
          com.docker.v2c.component.description=Detects\ Trusty\ Tahr
    CMD grep "PRETTY_NAME=\"Ubuntu 14.04.5 LTS\"" /v2c/disk/etc/os-release 1>2 2>/dev/null

#### Ubuntu 14.04.5 Runit Provisioner Contributed Dockerfile Fragment

    RUN apt-get update && apt-get install -y runit
    ENTRYPOINT ["runsvdir","-P","/etc/service"]
    STOPSIGNAL SIGHUP

You'll notice that the init detective for Ubuntu 14.04.5 is the same as the OS detective for the same version. However the important difference is that this detective is in the init category and points to ````runit-provisioner:ubuntu-v14.04.5````. That provisioner contributes a very simple Dockerfile fragment which installs runit for the target OS, sets the entrypoint, and sets the stop signal.

Runit stops on SIGHUP. That is important to note if you can't figure out why SIGINT fails to stop your container.

The most important part of the Dockerfile fragment is that the entrypoint tells runit to look for service definitions in /etc/services. Now any other packages that have been detected for the various source init systems out there can contribute runit service definitions to a known location.

#### Apache2 SysV Service Detective

    FROM alpine:3.4
    LABEL com.docker.v2c.component=detective \
          com.docker.v2c.component.category=init \
          com.docker.v2c.component.builtin=1 \
          com.docker.v2c.component.description=Detects\ SysV\ automatic\ startup\ for\ Apache2. \
          com.docker.v2c.component.rel=v2c/init.apache2-provisioner:2
    CMD ls -al /v2c/disk/etc/rc2.d/S91apache2 2>1 1>/dev/null || ls -al /v2c/disk/etc/rc3.d/S91apache2 2>1 1>/dev/null || ls -al /v2c/disk/etc/rc4.d/S91apache2 2>1 1>/dev/null || ls -al /v2c/disk/etc/rc5.d/S91apache2 2>1 1>/dev/null

This detective specifically detects Apache2 as commonly installed for System V init systems. Ubuntu 14.04.5 uses Upstart, but Apache2 packages are unaware of Upstart and so install the Apache2 service into System V runlevels pointing to a script in /etc/init.d.

#### Apache2 Runit Service Provisioner



