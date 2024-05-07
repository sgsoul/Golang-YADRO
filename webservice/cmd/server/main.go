package main

import (
	"github.com/sgsoul/internal/config"
	"github.com/sgsoul/internal/core/server"
)

func main() {
	cfg := config.New("config.yaml")

	go server.StartScheduler(23, 59)
	go server.StartServer(cfg)

	select {}
}