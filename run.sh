#!/bin/bash

set -xe

docker-compose down
docker-compose up --build -d
