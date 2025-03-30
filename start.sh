#!/usr/bin/env bash

docker build -t mitm-proxy . && docker run -p 8000:8000 -p 8080:8080 --name mitm-proxy --rm -it mitm-proxy
