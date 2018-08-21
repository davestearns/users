package main

import (
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"

	"github.com/caarlos0/env"
	"github.com/davestearns/sessions"
	"github.com/davestearns/userservice/handlers"
	"github.com/davestearns/userservice/models/users"
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

	//construct a new redis session store
	sessionStore := sessions.NewRedisStore(sessions.NewRedisPool(cfg.RedisAddr, time.Minute*10), time.Hour)

	//construct a new AWS session, and DynamoDB client
	awsSession, err := session.NewSession()
	if err != nil {
		log.Fatalf("error creating new AWS session: %v", err)
	}
	dynamoClient := dynamodb.New(awsSession)

	handlerConfig := &handlers.Config{
		SessionManager: sessions.NewManager(sessions.DefaultIDLength, cfg.SessionKeys, sessionStore),
		UserStore:      users.NewDynamoDBStore(dynamoClient, cfg.DynamoDBTable, cfg.DynamoDBKey),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/users", handlerConfig.UsersHandler)
	mux.HandleFunc("/users/", handlerConfig.EnsureSession(handlerConfig.SpecificUserHandler))

	log.Printf("server is listening at http://%s...", cfg.Addr)
	log.Fatal(http.ListenAndServe(cfg.Addr, mux))

}
