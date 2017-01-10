#!/bin/sh

# write a Dockerfile
printf "RUN echo this is a generated Dockerfile.\n" > /Dockerfile 2>&1
printf "RUN echo this is another command.\n" >> /Dockerfile 2>&1

tar-append Dockerfile
