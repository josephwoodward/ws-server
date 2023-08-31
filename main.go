package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	ws "github.com/josephwoodward/go-websocket-server/websocket"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", ws.Upgrade)
	server := &http.Server{Addr: ":8080", Handler: mux}

	go func() {
		fmt.Printf("starting server on port 8080")
		if err := server.ListenAndServe(); err != nil {
			fmt.Printf("failed to start server: %s", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("failed to shutdown server: %s", err)
	}
}
