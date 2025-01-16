package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/ChintuIdrive/idrive-e2-log-access/api"
	"github.com/gin-gonic/gin"
)

type AppConfig struct {
	AppLogDirMap map[string]string
}

func main() {
	appConfig := getConfig()
	logApi := api.NewLogApi(appConfig.AppLogDirMap)
	router := gin.Default()

	router.GET("/api/log/list-logs", logApi.ListLogFiles)
	router.GET("/api/log/download-log", logApi.DownloadLogFile)

	router.Run(":8888")
}

func getConfig() *AppConfig {
	appConfig := &AppConfig{}
	configFile, err := os.Open("app-config.json")
	if err != nil {
		log.Fatalf("Failed to open config file: %v", err)
	}
	defer configFile.Close()

	decoder := json.NewDecoder(configFile)
	err = decoder.Decode(&appConfig)
	if err != nil {
		log.Fatalf("Failed to decode config file: %v", err)
	}
	return appConfig
}
