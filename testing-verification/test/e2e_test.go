package e2e

import (
    "bytes"
    "fmt"
    "io"
    "net/http"
    "os"
    "testing"
    "time"

    "github.com/jasonlvhit/gocron"
    log "github.com/rs/zerolog/log"
    "github.com/sgsoul/internal/core"
    "github.com/sgsoul/internal/server"
    "github.com/sgsoul/internal/service"
    "github.com/sgsoul/internal/service/search"
    "github.com/sgsoul/internal/storage"
    "github.com/sgsoul/internal/xkcd"
    "github.com/stretchr/testify/assert"
    "gopkg.in/yaml.v2"
)

func Test_e2e(t *testing.T) {
    cfg := New("../config.yaml")
    if cfg == nil {
        t.Fatal("Failed to load configuration")
    }

	//db := storage.NewMySQLDB("root:12345@tcp(localhost:3306)/xkcd")
	db := storage.NewMySQLDB(cfg.DSN)
	sr := search.NewSearch(db)
	cl := xkcd.NewClient(cfg.SourceURL, db)
	src := service.NewService(cfg, db, cl)

	go StartScheduler(t)
	go server.StartServer(cfg, src, sr)

	select {}
}

func StartScheduler(t any) {
    now := time.Now()
    futureTime := now.Add(20 * time.Second)
    err := gocron.Every(1).Minute().From(&futureTime).Do(JustTest, t)
    if err != nil {
        log.Error().Err(err).Msg("Failed to schedule job")
        return
    }

    <-gocron.Start()
}

func JustTest(t *testing.T) {
    log.Info().Msg("next check")

    check := func() error {
        fmt.Println("Checking search results...")
        resp, err := http.Get("http://localhost:8080/pics?search=apple+doctor")
        if err != nil {
            return err
        }
        defer resp.Body.Close()
        body, err := io.ReadAll(resp.Body)
        if err != nil {
            return err
        }
        assert.Contains(t, string(body), "apple a day")
        return nil
    }

    log.Info().Msg("next upd")

    update := func() error {
        fmt.Println("Updating...")
        resp, err := http.Post("http://localhost:8080/update", "", nil)
        if err != nil {
            return err
        }
        defer resp.Body.Close()
        return nil
    }

    log.Info().Msg("next get")

    get := func() error {
        fmt.Println("Getting data...")
        resp, err := http.Get("http://localhost:8080/pics?search=apple+doctor")
        if err != nil {
            return err
        }
        defer resp.Body.Close()
        body, err := io.ReadAll(resp.Body)
        if err != nil {
            return err
        }
        fmt.Println(string(body))
        return nil
    }

    log.Info().Msg("next admin")

    admin := func() error {
        fmt.Println("Admin login...")
        data := []byte(`{"username":"admin","password":"adminpassword"}`)
        resp, err := http.Post("http://localhost:8080/login", "application/json", bytes.NewBuffer(data))
        if err != nil {
            return err
        }
        defer resp.Body.Close()
        return nil
    }

    // каналы для синхронизации
    done := make(chan struct{})
    defer close(done)

    go func() {
        steps := []func() error{admin, update, get, check}
        for _, step := range steps {
            log.Info().Msg("one more step")
            if err := step(); err != nil {
                log.Error().Err(err).Msg("Step failed")
            }
        }
        done <- struct{}{}
    }()

    <-done
}

func LoadConfig(configPath string) (*core.Config, error) {
    data, err := os.ReadFile(configPath)
    if err != nil {
        log.Error().Err(err).Msg("error reading configuration fail")
        return nil, err
    }

    config := &core.Config{}

    err = yaml.Unmarshal(data, config)
    if err != nil {
        log.Error().Err(err).Msg("error unmarshal configuration fail")
        return nil, err
    }

    return config, err
}

func New(configPath string) *core.Config {
    cfg, err := LoadConfig(configPath)
    if err != nil {
        log.Error().Err(err).Msg("error reading configuration fail")
        return nil
    }
    return cfg
}
