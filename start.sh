#!/usr/bin/env bash

docker build -t https-proxy .
docker run -p 8000:8000 -p 8080:8080 https-proxy
