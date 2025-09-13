package main

import (
	"database/sql"
	"exchrates/internal/api"
	"exchrates/internal/provider"
	"exchrates/internal/service"
	"exchrates/internal/store"
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

type Config struct {
	DBDriver   string
	DBUser     string
	DBPassword string
	DBName     string
	DBHost     string
	DBPort     string
	RSSURL     string
	LogFile    string
	ServerPort string
}

func NewConfig() *Config {
	getEnv := func(key, defaultValue string) (value string) {
		if val, exists := os.LookupEnv(key); exists && val != "" {
			value = val
		} else {
			value = defaultValue
		}
		if value == "" {
			log.Fatalf("missing required environment variable: %s", key)
		}
		return value
	}
	return &Config{
		DBDriver:   getEnv("DB_DRIVER", "mysql"),
		DBUser:     getEnv("DB_USER", ""),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", ""),
		DBHost:     getEnv("DB_HOST", ""),
		DBPort:     getEnv("DB_PORT", "3306"),
		RSSURL:     getEnv("RSS_URL", "https://www.bank.lv/vk/ecb_rss.xml"),
		LogFile:    getEnv("LOG_FILE", "./logs/log.txt"),
		ServerPort: getEnv("SERVER_PORT", ":8080"),
	}
}

func init() {
	os.Setenv("TZ", "Europe/Riga")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("No command provided.")
		fmt.Println(UsageTxt)
		os.Exit(1)
	}

	config := NewConfig()

	logFile, err := os.OpenFile(config.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	customLogger := logger.New(os.Stdout, logFile)
	customLogger.EnableDebug(true)

	connstr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", config.DBUser, config.DBPassword, config.DBHost, config.DBPort, config.DBName)

	db, err := sql.Open(config.DBDriver, connstr)
	if err != nil {
		customLogger.Fatal(err, "Failed to connect to database")
	}
	defer db.Close()

	provider := provider.NewRSSProvider(config.RSSURL)
	store, err := store.NewSqlDB(db)
	if err != nil {
		customLogger.Fatal(err, "Failed to connect to database")
	}
	svc := service.NewRateService(provider, store, customLogger)

	cmd := os.Args[1]
	switch cmd {
	case "server":
		runServer(svc, config.ServerPort)
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
	log.Println("Rates fetched and saved.")
}
