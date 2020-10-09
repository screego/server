package ws

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/screego/server/ws/outgoing"
	"net"
	"net/http"
	"strings"
	"time"
)

var ping = func(conn *websocket.Conn) error {
	return conn.WriteMessage(websocket.PingMessage, nil)
}

var writeJSON = func(conn *websocket.Conn, v interface{}) error {
	return conn.WriteJSON(v)
}

const (
	writeWait = 2 * time.Second
)

type Client struct {
	conn *websocket.Conn
	info ClientInfo
	once once
	read chan<- ClientMessage
}

type ClientMessage struct {
	Info     ClientInfo
	Incoming Event
}

type ClientInfo struct {
	ID                xid.ID
	RoomID            string
	Authenticated     bool
	AuthenticatedUser string
	Write             chan outgoing.Message
	Close             chan string
	Addr              net.IP
}

func newClient(conn *websocket.Conn, req *http.Request, read chan ClientMessage, authenticatedUser string, authenticated, trustProxy bool) *Client {

	ip := conn.RemoteAddr().(*net.TCPAddr).IP
	if realIP := req.Header.Get("X-Real-IP"); trustProxy && realIP != "" {
		ip = net.ParseIP(realIP)
	}

	client := &Client{
		conn: conn,
		info: ClientInfo{
			Authenticated:     authenticated,
			AuthenticatedUser: authenticatedUser,
			ID:                xid.New(),
			RoomID:            "",
			Addr:              ip,
			Write:             make(chan outgoing.Message, 1),
			Close:             make(chan string, 1),
		},
		read: read,
	}
	client.debug().Msg("WebSocket New Connection")
	conn.SetCloseHandler(func(code int, text string) error {
		message := websocket.FormatCloseMessage(code, text)
		client.debug().Str("reason", text).Int("code", code).Msg("WebSocket Close")
		return conn.WriteControl(websocket.CloseMessage, message, time.Now().Add(writeWait))
	})
	return client
}

// Close closes the connection.
func (c *Client) Close() {
	c.once.Do(func() {
		c.conn.Close()
		go func() {
			c.read <- ClientMessage{
				Info:     c.info,
				Incoming: &Disconnected{},
			}
		}()
	})
}

// startWriteHandler starts listening on the client connection. As we do not need anything from the client,
// we ignore incoming messages. Leaves the loop on errors.
func (c *Client) startReading(pongWait time.Duration) {
	defer c.Close()
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(appData string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		t, m, err := c.conn.NextReader()
		if err != nil {
			c.printWebSocketError("read", err)
			return
		}
		if t == websocket.BinaryMessage {
			_ = c.conn.CloseHandler()(websocket.CloseUnsupportedData, fmt.Sprintf("unsupported binary message type: %s", err))
			return
		}

		incoming, err := ReadTypedIncoming(m)
		if err != nil {
			_ = c.conn.CloseHandler()(websocket.CloseNormalClosure, fmt.Sprintf("malformed message: %s", err))
			return
		}
		c.debug().Interface("event", fmt.Sprintf("%T", incoming)).Msg("WebSocket Receive")
		c.read <- ClientMessage{Info: c.info, Incoming: incoming}
	}
}

// startWriteHandler starts the write loop. The method has the following tasks:
// * ping the client in the interval provided as parameter
// * write messages send by the channel to the client
// * on errors exit the loop
func (c *Client) startWriteHandler(pingPeriod time.Duration) {
	pingTicker := time.NewTicker(pingPeriod)

	dead := false
	conClosed := func() {
		dead = true
		c.Close()
		pingTicker.Stop()
	}
	defer conClosed()
	defer func() {
		c.debug().Msg("WebSocket Done")
	}()
	for {
		select {
		case reason := <-c.info.Close:
			if reason == CloseDone {
				return
			} else {
				_ = c.conn.CloseHandler()(websocket.CloseNormalClosure, reason)
				conClosed()
			}
		case message := <-c.info.Write:
			if dead {
				c.debug().Msg("WebSocket write on dead connection")
				continue
			}

			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			typed, err := ToTypedOutgoing(message)
			c.debug().Interface("event", typed.Type).Msg("WebSocket Send")
			if err != nil {
				c.debug().Err(err).Msg("could not get typed message, exiting connection.")
				conClosed()
				continue
			}

			if room, ok := message.(outgoing.Room); ok {
				c.info.RoomID = room.ID
			}

			if err := writeJSON(c.conn, typed); err != nil {
				conClosed()
				c.printWebSocketError("write", err)
			}
		case <-pingTicker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ping(c.conn); err != nil {
				conClosed()
				c.printWebSocketError("ping", err)
			}
		}
	}
}

func (c *Client) debug() *zerolog.Event {
	return log.Debug().Str("id", c.info.ID.String()).Str("ip", c.info.Addr.String())
}

func (c *Client) printWebSocketError(typex string, err error) {
	if strings.Contains(err.Error(), "use of closed network connection") {
		return
	}
	closeError, ok := err.(*websocket.CloseError)

	if ok && closeError != nil && (closeError.Code == 1000 || closeError.Code == 1001) {
		// normal closure
		return
	}

	c.debug().Str("type", typex).Err(err).Msg("WebSocket Error")
}
