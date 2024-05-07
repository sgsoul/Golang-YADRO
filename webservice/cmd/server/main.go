package main

import (
	"fmt"
	"net/http"
	"time"


	log "github.com/rs/zerolog/log"
	"github.com/jasonlvhit/gocron"
	"github.com/sgsoul/pkg/config"
	"github.com/sgsoul/pkg/server"
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Error().Err(err).Msg("Error loading configuration")
	}

	go startScheduler()
	go startServer(cfg)

	select {}
}

func startScheduler() {
	t := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 19, 57, 0, 0, time.FixedZone("UTC+3", 3*60*60))
	gocron.Every(1).Hour().From(&t).Do(update)

	<-gocron.Start()
}

func startServer(cfg *config.Config) {
	xkcdServer, err := server.NewServer(cfg)
	if err != nil {
		log.Error().Err(err).Msg("Error creating server")
	}

	if err := xkcdServer.Start(); err != nil {
		log.Error().Err(err).Msg("Error starting server")
	}
}

func update() {

	log.Info().Msg("Sheduled database update")
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Error().Err(err).Msg("Error loading configuration")
	}

	url := fmt.Sprintf("http://localhost:%d/update", cfg.Port)
	resp, err := http.Post(url, "", nil)
	if err != nil {
		log.Error().Err(err).Msg("Error executing request")
		return
	}
	resp.Body.Close()
}
