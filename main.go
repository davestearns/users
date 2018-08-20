package main

import (
	"log"
	"net/http"
	"time"

	"github.com/caarlos0/env"
	"github.com/davestearns/sessions"
	"github.com/davestearns/userservice/handlers"
)

type config struct {
	Addr          string   `env:"ADDR" envDefault:":80"`
	RedisAddr     string   `env:"REDIS_ADDR" envDefault:"127.0.0.1:6379"`
	SessionKeys   []string `env:"SESSION_KEYS,required"`
	DynamoDBTable string   `env:"DYNAMODB_TABLE" envDefault:"users"`
	DynamoDBKey   string   `env:"DYNAMODB_KEY" envDefault:"userName"`
}

func main() {
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("error loading configuration: %v", err)
	}
	log.Printf("using the following configuration: %+v", cfg)

	sessionStore := sessions.NewRedisStore(sessions.NewRedisPool(cfg.RedisAddr, time.Minute*10), time.Hour)
	handlerConfig := &handlers.Config{
		Manager: sessions.NewManager(sessions.DefaultIDLength, cfg.SessionKeys, sessionStore),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/users", handlerConfig.UsersHandler)

	log.Printf("server is listening at http://%s...", cfg.Addr)
	log.Fatal(http.ListenAndServe(cfg.Addr, mux))

}
