#!/usr/bin/fish

./build.fish || exit 1

hline Run soargs-server
./soargs-server/soargs-server start soargs-server

