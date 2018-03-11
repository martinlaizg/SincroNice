package main

import (
	"sincronice/client"
	"sincronice/server"
)

func main() {
	server.Run()
	client.Run()
}
