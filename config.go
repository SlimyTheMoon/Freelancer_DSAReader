package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"sync"
)

// --- CONFIGURATION ---

type AppConfig struct {
	LogPath string `json:"log_path"`
}

// Global config with mutex for thread safety
var (
	currentConfig AppConfig
	configMutex   sync.RWMutex
)

func GetLogPath() string {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return currentConfig.LogPath
}

func SetLogPath(path string) {
	configMutex.Lock()
	currentConfig.LogPath = path
	configMutex.Unlock()
}

func LoadConfig() {
	file, err := os.Open("config.json")
	if err != nil {
		currentConfig.LogPath = ""
		return
	}
	defer file.Close()
	json.NewDecoder(file).Decode(&currentConfig)
	if currentConfig.LogPath == "" {
		currentConfig.LogPath = ""
	}
}

func SaveConfig() {
	file, err := os.Create("config.json")
	if err != nil {
		log.Printf("Error saving config: %v", err)
		return
	}
	defer file.Close()
	json.NewEncoder(file).Encode(&currentConfig)
}

func ValidatePath(path string) bool {
	trimmedPath := strings.Trim(path, " \"'")
	_, err := os.Stat(trimmedPath)
	return err == nil
}
