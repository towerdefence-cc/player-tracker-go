package main

import (
	"player-tracker-go/servers"
	"player-tracker-go/service"
)

func main() {
	service.InitPlayerCounter()

	servers.InitGrpc()
	// blocking
	servers.InitRabbitMq()
}
