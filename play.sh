#!/usr/bin/env bash

docker build -t proxy-pg -f Dockerfile.playground . && docker run --name mitm --rm -it proxy-pg 
