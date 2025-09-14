package main

import (
	"database/sql"
	"exchrates/internal/api"
	"exchrates/internal/provider"
	"exchrates/internal/service"
	"exchrates/internal/store"
	"exchrates/pkg/config"
	"exchrates/pkg/logger"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

const (
	UsageTxt = `Usage: exchrates [server|fetch]`
)

func init() {
	os.Setenv("TZ", "Europe/Riga")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("No command provided.")
		fmt.Println(UsageTxt)
		os.Exit(1)
	}

	config := config.NewConfig()
	// config.LoadEnv() - to load all env vars
	config.LoadEnvVar("DB_DRIVER", "mysql")
	config.LoadEnvVar("DB_USER", "")
	config.LoadEnvVar("DB_PASSWORD", "")
	config.LoadEnvVar("DB_NAME", "")
	config.LoadEnvVar("DB_HOST", "")
	config.LoadEnvVar("DB_PORT", "3306")
	config.LoadEnvVar("RSS_URL", "https://www.bank.lv/vk/ecb_rss.xml")
	config.LoadEnvVar("LOG_FILE", "./logs/log.txt")
	config.LoadEnvVar("SERVER_PORT", ":8080")

	logFile, err := os.OpenFile(config.MustString("LOG_FILE"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	customLogger := logger.New(os.Stdout, logFile)
	customLogger.EnableDebug(true)

	connstr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		config.MustString("DB_USER"),
		config.MustString("DB_PASSWORD"),
		config.MustString("DB_HOST"),
		config.MustString("DB_PORT"),
		config.MustString("DB_NAME"))

	db, err := sql.Open(config.MustString("DB_DRIVER"), connstr)
	if err != nil {
		customLogger.Fatal(err, "Failed to connect to database")
	}
	defer db.Close()

	provider := provider.NewRSSProvider(config.MustString("RSS_URL"))
	store, err := store.NewSqlDB(db)
	if err != nil {
		customLogger.Fatal(err, "Failed to connect to database")
	}
	svc := service.NewRateService(provider, store, customLogger)

	cmd := os.Args[1]
	switch cmd {
	case "server":
		runServer(svc, config.MustString("SERVER_PORT"))
	case "fetch":
		runFetch(svc)
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		fmt.Println(UsageTxt)
		os.Exit(1)
	}
}

func runServer(svc *service.RateService, port string) {
	handler := api.NewHandler(svc)

	http.HandleFunc("/latest", handler.GetLatest)
	http.HandleFunc("/history", handler.GetHistory)

	svc.Logger.Info("Listening on %v\n", port)
	svc.Logger.Fatal(http.ListenAndServe(port, nil), "Failed to start server")
}

func runFetch(svc *service.RateService) {
	err := svc.FetchAndSave()
	if err != nil {
		svc.Logger.Fatal(err, "Failed to fetch and save rates")
	}
	svc.Logger.Info("Rates fetched and saved.")
}
