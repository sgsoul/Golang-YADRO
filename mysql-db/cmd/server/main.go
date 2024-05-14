package main

import (
	"github.com/sgsoul/internal/config"
	"github.com/sgsoul/internal/core/server"
)

func main() {
	cfg := config.New("config.yaml")

	go server.StartScheduler(3, 0)
	go server.StartServer(cfg)

	select {}
}
