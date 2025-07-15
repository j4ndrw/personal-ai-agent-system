#!/bin/bash

docker network inspect paias >/dev/null 2>&1 || docker network create paias
docker compose down && docker compose up --build --remove-orphans -d
docker exec -it paias-server python3 main.py
