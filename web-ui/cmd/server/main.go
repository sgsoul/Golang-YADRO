package main

import (
	"github.com/sgsoul/internal/core"
	"github.com/sgsoul/internal/server"
	"github.com/sgsoul/internal/service"
	"github.com/sgsoul/internal/service/search"
	"github.com/sgsoul/internal/storage"
	"github.com/sgsoul/internal/xkcd"
)

func main() {
	cfg := core.New("config.yaml")

	db := storage.NewMySQLDB(cfg.DSN)
	sr := search.NewSearch(db)
	cl := xkcd.NewClient(cfg.SourceURL, db)
	src := service.NewService(cfg, db, cl)

	go service.StartScheduler(3, 0, cfg.Port)
	go server.StartServer(cfg, src, sr)

	select {}
}
