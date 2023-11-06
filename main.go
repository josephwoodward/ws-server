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
			head, err := wsRes.Read(2)
			if err != nil {
				fmt.Printf(err.Error())
			}

			frame.IsFragment = (head[0] & 0x80) == 0x00
			// https://datatracker.ietf.org/doc/html/rfc6455#section-11.8
			// frame.Opcode = head[0] & 0x0F
			frame.Opcode2 = ws.WsOpCode(head[0] & 0x0F)
			frame.Reserved = (head[0] & 0x70)
			frame.IsMasked = (head[1] & 0x80) == 0x80
			frame.Length = uint64(head[1] & 0x7F)

			// if frame.IsMasked {
			// 	frame.MaskingKey = ""

			// }

			// if frame.Length == 126 {
			// 	data, err := wsRes.Read(2)
			// 	if err != nil {
			// 		return frame, err
			// 	}
			// 	length = uint64(binary.BigEndian.Uint16(data))
			// } else if frame.Length == 127 {
			// 	data, err := wsRes.Read(8)
			// 	if err != nil {
			// 		return frame, err
			// 	}
			// 	length = uint64(binary.BigEndian.Uint64(data))
			// }

			mask, err := wsRes.Read(4)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}

			payload, err := wsRes.Read(int(frame.Length)) // possible data loss
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}

			for i := uint64(0); i < frame.Length; i++ {
				payload[i] ^= mask[i%4]
			}
			frame.Payload = payload
			frame.MaskingKey = mask

			switch frame.Opcode2 {
			case ws.WsPingMessage, ws.WsPongMessage, ws.WsCloseMessage:
				fmt.Print("ping")
			case ws.WsTextMessage:
				fmt.Printf("payload: %s", frame.Payload)
			default:
				break

			}
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
