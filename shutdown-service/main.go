package main

import (
	"log"
	"net/http"
	"os"
	"os/exec"
)

func shutdownHandler(w http.ResponseWriter, r *http.Request, authToken string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Basic authentication check
	if r.Header.Get("Authorization") != "Bearer "+authToken {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Run the shutdown command
	cmd := exec.Command("sudo", "shutdown", "now")
	err := cmd.Run()
	if err != nil {
		log.Printf("Shutdown failed: %v", err)
		http.Error(w, "Shutdown failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Shutting down..."))
}

func main() {
	authToken := os.Getenv("SHUTDOWN_TOKEN")
	if authToken == "" {
		log.Fatal("SHUTDOWN_TOKEN environment variable not set")
	}

	http.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		shutdownHandler(w, r, authToken)
	})

	log.Println("shutdown-service starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
