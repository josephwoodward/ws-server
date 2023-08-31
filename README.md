Web Socket Server

A simple Web Socket server written in Go used for learning more about the [Web Socket protocol](https://datatracker.ietf.org/doc/html/rfc6455#section-1.2).

Step 1: Establishing the Opening/Closing handshake

```
// Client

GET /chat HTTP/1.1
Host: server.example.com
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==
Origin: http://example.com
Sec-WebSocket-Protocol: chat, superchat
Sec-WebSocket-Version: 13
```

```
// Server

HTTP/1.1 101 Switching Protocols
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Accept: s3pPLMBiTxaQ9kYGzzhZRbK+xOo=
```
