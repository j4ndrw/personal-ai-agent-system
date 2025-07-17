#!/bin/bash


cd ./server && pipenv install && ./freeze_requirements.sh && cd ..
docker network inspect paias >/dev/null 2>&1 || docker network create paias
docker compose down && docker compose up --build --remove-orphans -d
docker logs paias-server -f
