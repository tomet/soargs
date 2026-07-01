#!/usr/bin/fish

hline Build soargs
go build . || exit 1

hline Build soargs-client
go build -o ./soargs-client/soargs-client ./soargs-client/ || exit 1

hline Build soargs-server
go build -o ./soargs-server/soargs-server/ ./soargs-server/ || exit 1
