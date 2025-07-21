#!/bin/bash

function startup_paias_server() {
    cd ./server && pipenv install && ./freeze_requirements.sh && cd ..
    docker network inspect paias >/dev/null 2>&1 || docker network create paias
    docker compose down && docker compose up --build --remove-orphans -d
}

function run_paias_client() {
    cd ./client && ./.build/paias
}

# function kill_paias_server() {
#     docker ps -aq --filter "name=paias" | xargs -r docker rm -f
#     exit 0
# }

startup_paias_server
sleep 2
run_paias_client
