package api

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"sync"

	"void/internal/event"
)

const wsMagicGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

// wsConn wraps a hijacked TCP connection after a completed WebSocket
// upgrade, exposing a minimal WriteText method (server -> client push only,
// which is all the Simulation Control Center's live dashboards need).
type wsConn struct {
	mu   sync.Mutex
	conn net.Conn
	bw   *bufio.Writer
}

func (c *wsConn) WriteText(payload []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	frame := encodeTextFrame(payload)
	if _, err := c.bw.Write(frame); err != nil {
		return err
	}
	return c.bw.Flush()
}

func (c *wsConn) Close() { _ = c.conn.Close() }

// encodeTextFrame builds a minimal unmasked RFC6455 text frame (server ->
// client frames are never masked per spec).
func encodeTextFrame(payload []byte) []byte {
	var header []byte
	n := len(payload)
	switch {
	case n <= 125:
		header = []byte{0x81, byte(n)}
	case n <= 65535:
		header = []byte{0x81, 126, byte(n >> 8), byte(n)}
	default:
		header = []byte{0x81, 127,
			byte(n >> 56), byte(n >> 48), byte(n >> 40), byte(n >> 32),
			byte(n >> 24), byte(n >> 16), byte(n >> 8), byte(n)}
	}
	return append(header, payload...)
}

// upgradeWebSocket performs the RFC6455 handshake over a hijacked
// connection and returns a wsConn ready for server-push writes.
func upgradeWebSocket(w http.ResponseWriter, r *http.Request) (*wsConn, error) {
	key := r.Header.Get("Sec-WebSocket-Key")
	hj, ok := w.(http.Hijacker)
	if !ok {
		return nil, errNotHijackable
	}
	conn, rw, err := hj.Hijack()
	if err != nil {
		return nil, err
	}
	accept := computeAcceptKey(key)
	resp := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + accept + "\r\n\r\n"
	if _, err := rw.Writer.WriteString(resp); err != nil {
		conn.Close()
		return nil, err
	}
	if err := rw.Writer.Flush(); err != nil {
		conn.Close()
		return nil, err
	}
	return &wsConn{conn: conn, bw: rw.Writer}, nil
}

func computeAcceptKey(key string) string {
	h := sha1.New()
	h.Write([]byte(key + wsMagicGUID))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

type simpleErr string

func (e simpleErr) Error() string { return string(e) }

const errNotHijackable = simpleErr("response writer does not support hijacking")

// ---- Subscriber registry: broadcasts tick/event updates to connected UIs ----

type wsHub struct {
	mu   sync.Mutex
	subs map[string][]*wsConn // simulationID -> connections
}

func newWSHub() *wsHub { return &wsHub{subs: map[string][]*wsConn{}} }

func (h *wsHub) add(simID string, c *wsConn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.subs[simID] = append(h.subs[simID], c)
}

func (h *wsHub) broadcast(simID string, payload []byte) {
	h.mu.Lock()
	conns := append([]*wsConn{}, h.subs[simID]...)
	h.mu.Unlock()
	for _, c := range conns {
		if err := c.WriteText(payload); err != nil {
			log.Printf("ws write error (sim %s): %v", simID, err)
		}
	}
}

// tickMessage is the JSON payload pushed to subscribed dashboards every tick.
type tickMessage struct {
	Type   string        `json:"type"`
	SimID  string        `json:"simulation_id"`
	Tick   int64         `json:"tick"`
	Events []event.Event `json:"events"`
}

func (a *App) broadcastTick(simID string, tick int64, events []event.Event) {
	if a.hub == nil {
		return
	}
	msg := tickMessage{Type: "tick", SimID: simID, Tick: tick, Events: events}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	a.hub.broadcast(simID, data)
}

func (a *App) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	simID := r.PathValue("id")
	conn, err := upgradeWebSocket(w, r)
	if err != nil {
		http.Error(w, "websocket upgrade failed: "+err.Error(), http.StatusBadRequest)
		return
	}
	if a.hub == nil {
		a.hub = newWSHub()
	}
	a.hub.add(simID, conn)
	_ = conn.WriteText([]byte(`{"type":"connected","simulation_id":"` + simID + `"}`))
	// Server-push only: we don't need to read client frames for this
	// use case, so the connection is kept alive purely by broadcastTick
	// calls; the goroutine exits naturally once the peer closes the socket
	// and a subsequent write fails.
}
