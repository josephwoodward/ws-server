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
	ws  *ws.WsUpgradeResult
}

func main() {
	mux := http.NewServeMux()
	const port = ":8082"
	server := &server{
		srv: &http.Server{Addr: port, Handler: mux},
	}

	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		wsRes, err := ws.Upgrade(w, r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		for {
			frame := ws.Frame{}
			head, err := wsRes.Read2(2)
			if err != nil {
				fmt.Printf(err.Error())
			}

			frame.IsFragment = (head[0] & 0x80) == 0x00
			// https: //datatracker.ietf.org/doc/html/rfc6455#section-11.8
			frame.Opcode = head[0] & 0x0F
			frame.Opcode2 = ws.WsOpCode(head[0] & 0x0F)
			frame.Reserved = (head[0] & 0x70)

			// switch frame.Opcode {
			switch frame.Opcode2 {
			case ws.WsPingMessage, ws.WsPongMessage, ws.WsCloseMessage:
				fmt.Print("ping")
			default:
				break

			}

			fmt.Print(head)

			_ = head
		}
		// server.ws = wsRes
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
