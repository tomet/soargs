#!/usr/bin/fish

hline Build soargs
go build . || exit 1

hline Build soargs-client
go build -C soargs-client . || exit 1

hline Build soargs-server
go build -C soargs-server . || exit 1
