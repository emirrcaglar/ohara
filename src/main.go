package main

import (
	"fmt"
	"net/http"

	"ohara/src/internal/server"
)

func main() {
	mangaDir := "C:/Users/emirc/Downloads"
	port := ":8080"

	// 2. Initialize the Server (Router)
	// We pass the mangaDir here so the internal package knows where to look
	router := server.New(mangaDir)

	// 3. Start the Server
	fmt.Printf("Ohara is running on http://0.0.0.0%s\n", port)
	fmt.Printf("Serving manga from: %s\n", mangaDir)

	// ListenAndServe blocks forever, so we wrap it in a log.Fatal to catch crashes
	err := http.ListenAndServe(port, router)
	if err != nil {
		fmt.Printf("Server failed to start: %v", err)
	}
}
