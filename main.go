package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-acme/lego/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/gorilla/mux"

	"github.com/buzzology/slack_bot/db"
	"github.com/buzzology/slack_bot/handlers"
	"github.com/buzzology/slack_bot/service"
	"github.com/buzzology/slack_bot/service/api"
)

var (
	wait                  = flag.Duration("graceful-timeout", time.Second*15, "the duration to wait before timing out")
	slackWebhooksEndpoint = flag.String("slack-webhooks-endpoint", "0.0.0.0:3000", "Slack Webhooks Endpoint")
	configFile            = flag.String("config", "./config/local.toml", "Path to the config file to use.")
)

func main() {
	flag.Parse()
	log := logrus.New()

	// Load environment variable overrides (if any).
	log.Println("Loading config file: " + *configFile)
	config := service.NewConfig()
	viper.SetConfigFile(*configFile)

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w \n", err))
	}

	viper.Unmarshal(&config)

	if err := run(log, config); err != nil {
		log.Fatalf("failed running slack webhooks server: %s", err)
	}
}

func run(
	log *logrus.Logger,
	config *service.Config,
) error {
	log.Println("Starting webhooks server: ", *slackWebhooksEndpoint)

	// Create a db connection.
	sqlDb, err := initDb(config.DbConnection)
	if err != nil {
		log.Fatalf("failed to initialise database. %v", err)
	}

	// Instantiate repos.
	channelAccountsRepo := db.NewChannelAccountsRepo(sqlDb, log)
	messageBountiesRepo := db.NewMessageBountiesRepo(sqlDb, log)
	botStateRepo := db.NewBotStateRepo(sqlDb, log)
	botMessagesRepo := db.NewBotMessagesRepo(sqlDb, log)

	// Instantiate api clients.
	slackApiClient := api.NewSlackApiClient(
		config.ApiConfig,
		log,
	)

	// Instantiate services.
	channelAccountsService := service.NewChannelAccountsService(config, channelAccountsRepo, *slackApiClient)
	botStateService := service.NewBotStateService(config, botStateRepo, channelAccountsRepo, channelAccountsService, *slackApiClient, log)
	botMessagesService := service.NewBotMessagesService(config, botMessagesRepo, *slackApiClient, log)

	// Start the bot state maintainer to ensure we reset trackers when required etc.
	go maintainBotState(botStateService, log)

	// Prepare handlers.
	handler := handlers.NewSlackBotHandler(
		config,
		log,
		slackApiClient,
		channelAccountsRepo,
		messageBountiesRepo,
		channelAccountsService,
		botMessagesService,
		botMessagesRepo,
	)

	// Create router and add routes.
	r := mux.NewRouter()
	r.HandleFunc("/", handler.WebhookHandler).Methods("POST")
	r.HandleFunc("/slash_commands", handler.SlashCommandHandler).Methods("POST")
	r.HandleFunc("/interactions", handler.InteractionsHandler).Methods("POST")
	r.HandleFunc("/ping", PingHandler)

	srv := &http.Server{
		Addr: *slackWebhooksEndpoint,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		log.Println("Listening...")

		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), *wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("shutting down")
	os.Exit(0)

	return nil
}

// initDb creates initialises the connection to mysql.
func initDb(connectionString string) (*sql.DB, error) {

	log.Infof("initialising db connection")

	var sqlDb, err = sql.Open("mysql", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %v", err)
	}

	// Ensure that the database can be reached
	err = sqlDb.Ping()
	if err != nil {
		return nil, fmt.Errorf("error on opening database connection: %s", err.Error())
	}

	return sqlDb, nil
}

func maintainBotState(botStateService *service.BotStateService, log *logrus.Logger) {
	// Perform an initial tickover and then just check every few minutes.
	err := botStateService.Tickover(context.Background(), time.Now())
	if err != nil {
		log.WithError(err).Error("Failed to perform an initial tickover when starting maintain.")
		return
	}

	// https://stackoverflow.com/a/40364927/522859
	ticker := time.NewTicker(5 * time.Minute)
	for {
		select {
		case <-ticker.C:
			botStateService.Tickover(context.Background(), time.Now())
		}
	}
}

func PingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Pong")
}
