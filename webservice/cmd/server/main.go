package main

import (
	"log"

	//"github.com/jasonlvhit/gocron"
	"github.com/sgsoul/pkg/config"
	"github.com/sgsoul/pkg/server"
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// t := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 20, 38, 0, 0, time.FixedZone("UTC+3", 8*60*60))
	// gocron.Every(1).Hour().From(&t).Do(update)

	// <- gocron.Start()

	// fmt.Print("hhe")

	xkcdServer, _ := server.NewServer(cfg)
	if err := xkcdServer.Start(); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

// func update() {
// 	cfg, err := config.LoadConfig("config.yaml")
// 	if err != nil {
// 		log.Fatalf("Error loading configuration: %v", err)
// 	}

// 	url := fmt.Sprintf("http://localhost:%d/update", cfg.Port)
// 		resp, err := http.Post(url, "", nil)
// 		if err != nil {
// 			log.Fatal("Ошибка при выполнении запроса:", err)
// 			return
// 		}
// 		resp.Body.Close()
// }