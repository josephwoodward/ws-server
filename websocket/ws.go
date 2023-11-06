package ws

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
)

// type WebSocket struct {
// 	conn      net.Conn
// 	bufrw     *bufio.ReadWriter
// 	header    http.Header
// 	something int
// }

type WsUpgradeResult struct {
	conn  net.Conn
	bufrw *bufio.ReadWriter
}

// func (ws *WsUpgradeResult) ReadLoop() {
// 	for {
// 		frame := Frame{}
// 		head, err := ws.Read(2)
// 		if err != nil {
// 			fmt.Printf(err.Error())
// 		}

// 		frame.IsFragment = (head[0] & 0x80) == 0x00
// 		frame.Opcode = head[0] & 0x0F
// 		frame.Reserved = (head[0] & 0x70)
// 	}
// }

func (ws *WsUpgradeResult) Read2(sz int) ([]byte, error) {
	data := make([]byte, 4096)
	for {
		bytesRead, err := ws.bufrw.Read(data)
		if err != nil && err != io.EOF {
			return data, err
		}
		if bytesRead > 0 {
			break
		}
	}

	return data, nil
}

func (ws *WsUpgradeResult) Read3() ([]byte, error) {
	data := make([]byte, 4096)
	_, err := ws.bufrw.Read(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (ws *WsUpgradeResult) Read(sz int) ([]byte, error) {
	data := make([]byte, 0)
	for {
		if len(data) == sz {
			break
		}
		// Temporary slice to read chunk
		sz := 4096
		remaining := sz - len(data)
		if sz > remaining {
			sz = remaining
		}
		temp := make([]byte, sz)

		n, err := ws.bufrw.Read(temp)
		if err != nil && err != io.EOF {
			fmt.Print("error")
			return data, err
		}

		data = append(data, temp[:n]...)
		// fmt.Printf("contents: %s", data)
		fmt.Println(data) // AB
	}

	return data, nil
}

func GenerateAcceptHash(key string) string {
	h := sha1.New()
	h.Write([]byte(key))
	h.Write([]byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func Upgrade(w http.ResponseWriter, r *http.Request) (*WsUpgradeResult, error) {
	// Opening handshake: https://datatracker.ietf.org/doc/html/rfc6455#section-1.3
	// GET /chat HTTP/1.1
	// Host: server.example.com
	// Upgrade: websocket
	// Connection: Upgrade
	// Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==
	// Origin: http://example.com
	// Sec-WebSocket-Protocol: chat, superchat
	// Sec-WebSocket-Version: 13
	if r.Method != "GET" {
		http.Error(w, "request method must be a GET", http.StatusMethodNotAllowed)
		return nil, fmt.Errorf("request method must be a GET")
	}

	if h := r.Header.Get("Upgrade"); h != "websocket" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		http.Error(w, "invalid value for header 'upgrade'", http.StatusMethodNotAllowed)
	}

	if h := r.Header.Get("Connection"); h != "Upgrade" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		http.Error(w, "invalid value for header 'connection'", http.StatusMethodNotAllowed)
	}

	var key string
	if key = r.Header.Get("Sec-WebSocket-Key"); key == "" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		http.Error(w, "invalid value for header 'connection'", http.StatusMethodNotAllowed)
	}

	hj, ok := w.(http.Hijacker)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
	}

	conn, bufrw, err := hj.Hijack()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		conn.Close()
	}

	if bufrw.Reader.Buffered() > 0 {
		conn.Close()
		return nil, fmt.Errorf("client data was sent before handshake completed")
	}

	var buf [1024]byte
	p := buf[:0]

	// HTTP/1.1 101 Switching Protocols
	// Upgrade: websocket
	// Connection: Upgrade
	// Sec-WebSocket-Accept: s3pPLMBiTxaQ9kYGzzhZRbK+xOo=
	// From https://tools.ietf.org/html/rfc6455#section-4.2.2
	p = append(p, "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: "...)
	p = append(p, GenerateAcceptHash(key)...)
	p = append(p, "\r\n"...)
	p = append(p, "\r\n"...)

	if _, err = bufrw.Write(p); err != nil {
		return nil, err
	}

	if err = bufrw.Flush(); err != nil {
		return nil, err
	}

	ws := &WsUpgradeResult{
		conn:  conn,
		bufrw: bufrw,
	}

	return ws, nil
}
