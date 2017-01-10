#!/bin/sh

# write a Dockerfile
printf "RUN echo this is a generated Dockerfile from app.random2.\n" > /Dockerfile 2>&1
printf "RUN echo this is another command from app.random2.\n" >> /Dockerfile 2>&1

tar-append Dockerfile
