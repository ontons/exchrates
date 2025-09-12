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
	logger.InitFile(config.LogFile)

	connstr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", config.DBUser, config.DBPassword, config.DBHost, config.DBPort, config.DBName)

	db, err := sql.Open(config.DBDriver, connstr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	provider := provider.NewRSSProvider(config.RSSURL)
	store := store.NewSqlDB(db)
	svc := service.NewRateService(provider, store)

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

	log.Printf("Listening on %v\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func runFetch(svc *service.RateService) {
	err := svc.FetchAndSave()
	if err != nil {
		logger.Debug(err.Error())
		os.Exit(1)
	}
	log.Println("Rates fetched and saved.")
}
