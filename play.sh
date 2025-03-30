#!/usr/bin/env bash

docker build -t mitm-proxy-pg -f Dockerfile.playground . && docker run --name mitm-proxy-pg --rm -it mitm-proxy-pg
