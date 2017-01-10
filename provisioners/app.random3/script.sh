#!/bin/sh

# write a Dockerfile
printf "# No additional commands required for app.random3\n" >> /Dockerfile 2>&1

tar-append Dockerfile
