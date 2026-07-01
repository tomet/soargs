module github.com/tomet/soargs-server

go 1.26.1

replace github.com/tomet/soargs => ..

replace github.com/tomet/terror => ../../terror

require (
	github.com/tomet/cmdr v0.2.4
	github.com/tomet/soargs v0.0.0-20260621122004-e9bbf2247c6c
	github.com/tomet/terror v0.1.7
)
