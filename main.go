package main

import (
	"log"
	"net/http"
	"time"

	"github.com/caarlos0/env"
	"github.com/davestearns/sessions"
	"github.com/davestearns/users/handlers"
)

type config struct {
	addr          string   `env:"ADDR" envDefault:":80"`
	redisAddr     string   `env:"REDIS_ADDR" envDefault:"127.0.0.1:6379"`
	sessionKeys   []string `env:"SESSION_KEYS,required"`
	dynamoDBTable string   `env:"DYNAMODB_TABLE" envDefault:"users"`
	dynamoDBKey   string   `env:"DYNAMODB_KEY" envDefault:"userName"`
}

func main() {
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("error loading configuration: %v", err)
	}
	log.Printf("using the following configuration: %+v", cfg)

	handlerConfig := &handlers.Config{
		Manager: sessions.NewManager(sessions.DefaultIDLength,
			cfg.sessionKeys,
			sessions.NewRedisStore(sessions.NewRedisPool(cfg.redisAddr, time.Minute*10), time.Hour)),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/users", handlerConfig.UsersHandler)

	log.Printf("server is listening at http://%s...", cfg.addr)
	log.Fatal(http.ListenAndServe(cfg.addr, mux))

}
