package main

import (
	"fmt"
	"log"
	"net/http"
)

// --- CONSTANTS ---

const (
	Port     = ":8443"
	CertFile = "server.crt"
	KeyFile  = "server.key"
)

// --- UTILS ---

// SetupRoutes registers all HTTP routes
func SetupRoutes() {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/stream", handleStream)
	http.HandleFunc("/config", handleConfig)
	http.HandleFunc("/download", handleDownload)
}

// handleIndex serves the main HTML page
func handleIndex(w http.ResponseWriter, r *http.Request) {
	HandleIndex(w, r)
}

// handleStream handles SSE streaming
func handleStream(w http.ResponseWriter, r *http.Request) {
	HandleStream(w, r)
}

// handleConfig handles configuration updates
func handleConfig(w http.ResponseWriter, r *http.Request) {
	HandleConfig(w, r)
}

// handleDownload handles log file downloads
func handleDownload(w http.ResponseWriter, r *http.Request) {
	HandleDownload(w, r)
}

// StartServer starts the HTTP server
func StartServer() {
	fmt.Println("----------------------------------------------------------------")
	fmt.Printf("Freelancer Log Reader Active\n")
	fmt.Printf("Web Interface: https://localhost%s\n", Port)
	if GetLogPath() != "" {
		fmt.Printf("Current Log: %s\n", GetLogPath())
	}
	fmt.Println("----------------------------------------------------------------")

	err := http.ListenAndServeTLS(Port, CertFile, KeyFile, nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
