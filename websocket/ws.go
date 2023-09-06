package ws

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
)

type WebSocket struct {
	conn      net.Conn
	bufrw     *bufio.ReadWriter
	header    http.Header
	something int
}

type WsUpgradeResult struct {
	conn net.Conn
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
	if key = r.Header.Get("Sec-Websocket-Key"); key == "" {
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
	p = append(p, generateAcceptHash(key)...)

	if _, err = conn.Write(p); err != nil {
		conn.Close()
		return nil, err
	}

	result := &WsUpgradeResult{
		conn: conn,
	}

	return result, nil
}

func generateAcceptHash(key string) string {
	h := sha1.New()
	h.Write([]byte(key))
	h.Write([]byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
