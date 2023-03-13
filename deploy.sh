#!/bin/bash
docker build . -t xbellum:latest
docker save --output xbellum.tar xbellum:latest
rsync -avz xbellum.tar mirepoi-2:~
ssh mirepoi-2 "sudo docker load --input xbellum.tar"
# ssh charlie "cd /data/traefik; sudo docker-compose up -d blog;"
