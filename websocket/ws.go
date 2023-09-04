package ws

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"net"
	"net/http"
)

type WebSocket struct {
	conn      net.Conn
	bufrw     *bufio.ReadWriter
	header    http.Header
	something int
}

func Upgrade(w http.ResponseWriter, r *http.Request) {
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
	if key = r.Header.Get("Sec-Websocket-Key"); key == "" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		http.Error(w, "invalid value for header 'connection'", http.StatusMethodNotAllowed)
	}

	// TODO: Response
	// HTTP/1.1 101 Switching Protocols
	// Upgrade: websocket
	// Connection: Upgrade
	// Sec-WebSocket-Accept: s3pPLMBiTxaQ9kYGzzhZRbK+xOo=
	hj, ok := w.(http.Hijacker)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
	}
	conn, bufrw, err := hj.Hijack()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	_ = &WebSocket{conn, bufrw, r.Header, 1000}

	w.Header().Add("Upgrade", "websocket")
	w.Header().Add("Connection", "Upgrade")
	w.Header().Add("Sec-WebSocket-Accept", getAcceptHash(key))
	w.WriteHeader(http.StatusSwitchingProtocols)
}

func getAcceptHash(key string) string {
	h := sha1.New()
	h.Write([]byte(key))
	h.Write([]byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
