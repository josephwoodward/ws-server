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

type server struct {
	srv *http.Server
	// ws  *ws.WebSocket
	ws *ws.WsUpgradeResult
}

func main() {
	mux := http.NewServeMux()
	const port = ":8081"
	server := &server{
		srv: &http.Server{Addr: ":8081", Handler: mux},
	}

	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		res, err := ws.Upgrade(w, r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		server.ws = res
	})

	go func() {
		fmt.Printf("starting server on port %s\n", port)
		if err := server.srv.ListenAndServe(); err != nil {
			fmt.Printf("failed to start server: %s\n", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.srv.Shutdown(ctx); err != nil {
		fmt.Printf("failed to shutdown server: %s\n", err)
	}
}
