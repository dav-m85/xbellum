#!/bin/bash
docker build . -t xbellum:latest
docker save --output xbellum.tar xbellum:latest
rsync -avz xbellum.tar mirepoi-2:~
ssh mirepoi-2 "sudo docker load --input xbellum.tar"
rm xbellum.tar
ssh mirepoi-2 "cd /data; docker-compose -f dav-compose.yml up --force-recreate -d;"
