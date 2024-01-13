package main

import (
	"context"
	"fmt"
	"io"
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

			// Read initial 8 bits (2 bytes). This will give us enough information to decide how we handle the frame.
			head := make([]byte, 2)
			if _, err = wsRes.Bufrw.Read(head); err != nil {
				fmt.Print("breaking head read")
				break
			}

			if len(head) == 0 {
				continue
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

			frame.IsFinal = (head[0] & 0x80) == 0x00
			frame.Opcode = ws.WsOpCode(head[0] & 0x0F)
			fmt.Printf("For byte '%d', Opcode is %d\n", head[0], frame.Opcode)
			frame.Reserved = (head[0] & 0x70)

			frame.IsMasked = (head[1] & 0x80) == 0x80
			frame.Length = uint64(head[1] & 0x7F)

			maskingKey := make([]byte, 4)
			if _, err = wsRes.Bufrw.Read(maskingKey); err == io.EOF {
				fmt.Print("breaking masking key")
				break
			}

			frame.MaskingKey = maskingKey

			payload := make([]byte, int(frame.Length))
			if _, err = wsRes.Bufrw.Read(payload); err == io.EOF {
				fmt.Print("breaking payload length")
				break
			}

			for i := uint64(0); i < frame.Length; i++ {
				payload[i] ^= frame.MaskingKey[i%4]
			}
			frame.Payload = payload

			switch frame.Opcode {
			case ws.WsCloseMessage:
				fmt.Print("closing connection")
			case ws.WsPingMessage, ws.WsPongMessage:
				fmt.Print("ping / pong")
			case ws.WsTextMessage:
				fmt.Printf("received payload message: %s\n", string(frame.Payload))
				f := ws.Frame{
					IsFinal:  true,
					Opcode:   ws.WsTextMessage,
					IsMasked: false,
					Payload:  []byte("Hello Mike"),
				}
				wsRes.Write(f)
			default:
				break
			}
		}
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
