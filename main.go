package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/secretsmanager"

	"github.com/caarlos0/env"
	"github.com/davestearns/sessions"
	"github.com/davestearns/userservice/handlers"
	"github.com/davestearns/userservice/models/users"
)

type config struct {
	Addr          string   `env:"ADDR" envDefault:":80"`
	RedisAddr     string   `env:"REDIS_ADDR" envDefault:"cache.info441.info:6379"`
	SessionKeys   []string `env:"SESSION_KEYS"`
	DynamoDBTable string   `env:"DYNAMODB_TABLE" envDefault:"users"`
	DynamoDBKey   string   `env:"DYNAMODB_KEY" envDefault:"userName"`
}

func fetchSigningKeys(awsSession *session.Session) ([]string, error) {
	secretsClient := secretsmanager.New(awsSession)
	result, err := secretsClient.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: aws.String("userservice/sessionkeys"),
	})
	if err != nil {
		return nil, fmt.Errorf("error getting session signing keys: %v", err)
	}
	if result.SecretString == nil || len(*result.SecretString) == 0 {
		return nil, fmt.Errorf("session signing keys were empty")
	}
	return strings.Split(*result.SecretString, ","), nil
}

func main() {
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("error loading configuration: %v", err)
	}
	log.Printf("using the following configuration: %+v", cfg)

	//create a new AWS session
	awsSession, err := session.NewSession()
	if err != nil {
		log.Fatalf("error creating new AWS session: %v", err)
	}

	//if session keys are not supplied via the environment
	//get them from the AWS secrets service
	if len(cfg.SessionKeys) == 0 {
		keys, err := fetchSigningKeys(awsSession)
		if err != nil {
			log.Fatalf("error getting session signing keys: %v", err)
		}
		log.Printf("successfully fetched session keys from secrets service")
		cfg.SessionKeys = keys
	}

	//construct a new DynamoDB client
	dynamoClient := dynamodb.New(awsSession)

	//construct a new redis session store
	sessionStore := sessions.NewRedisStore(sessions.NewRedisPool(cfg.RedisAddr, time.Minute*10), time.Hour)

	handlerConfig := &handlers.Config{
		SessionManager: sessions.NewManager(sessions.DefaultIDLength, cfg.SessionKeys, sessionStore),
		UserStore:      users.NewDynamoDBStore(dynamoClient, cfg.DynamoDBTable, cfg.DynamoDBKey),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.RootHandler)
	mux.HandleFunc("/users", handlerConfig.UsersHandler)
	mux.HandleFunc("/users/", handlerConfig.EnsureSession(handlerConfig.SpecificUserHandler))
	mux.HandleFunc("/sessions", handlerConfig.SessionsHandler)
	mux.HandleFunc("/sessions/mine", handlerConfig.SessionsMineHandler)

	log.Printf("server is listening at http://%s...", cfg.Addr)
	log.Fatal(http.ListenAndServe(cfg.Addr, mux))

}
