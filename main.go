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

			// read first 2 bytes (16 bits)
			head, err := wsRes.Read(2)
			if err != nil {
				fmt.Printf(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
			}

			// https://datatracker.ietf.org/doc/html/rfc6455#section-5.2
			// 1 byte (8 bits) is the smallest addressable unit of memory
			// first 4 bits are flags with 3 reserved
			// first 1 bit is Fragment, using 128 binary representation as a bool flag via bitwise operator
			// remaing 4 bits are opcode
			// 129 = 1000 0001

			// Fragment is first bit so target first bit to determind fragment
			// 10000000 is 128 decimal, 0x80 hexidecimal
			// 0x00 = 0000000
			// fmt.Printf("The value is: %t", (10000001&10000000) == 00000000) = false
			// 10000001
			// 10000000 = 0

			frame.IsFragment = (head[0] & 0x80) == 0x00
			frame.Opcode = ws.WsOpCode(head[0] & 0x0F)
			frame.Reserved = (head[0] & 0x70)

			frame.IsMasked = (head[1] & 0x80) == 0x80
			frame.Length = uint64(head[1] & 0x7F)

			frame.MaskingKey, err = wsRes.Read(4)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}

			payload, err := wsRes.Read(int(frame.Length)) // possible data loss
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}

			for i := uint64(0); i < frame.Length; i++ {
				payload[i] ^= frame.MaskingKey[i%4]
			}
			frame.Payload = payload

			switch frame.Opcode {
			case ws.WsPingMessage, ws.WsPongMessage, ws.WsCloseMessage:
				fmt.Print("ping")
			case ws.WsTextMessage:
				fmt.Printf("payload: %s", frame.Payload)
				f := ws.Frame{
					IsFragment: false,
					Opcode:     ws.WsTextMessage,
					IsMasked:   false,
					Payload:    []byte("Hello Mike"),
				}
				wsRes.Write(f)
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
