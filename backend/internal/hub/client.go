package hub

import (
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait = 10 * time.Second
	pongWait  = 60 * time.Second
	pingEvery = 30 * time.Second
)

type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	serverID string
}

// Serve registers a browser connection and starts its two goroutines.
func (h *Hub) Serve(conn *websocket.Conn, serverID string) {
	c := &Client{hub: h, conn: conn, send: make(chan []byte, 16), serverID: serverID}
	h.subscribe(c)
	go c.writePump()
	go c.readPump()
}

// writePump sends metrics to the browser, plus a ping every 30 seconds.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingEvery)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok { // the hub removed this client
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump notices when the browser closes the tab or loses connection.
func (c *Client) readPump() {
	defer func() {
		c.hub.unsubscribe(c)
		c.conn.Close()
	}()
	c.conn.SetReadLimit(512)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			return
		}
	}
}
