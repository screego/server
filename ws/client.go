package ws

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/screego/server/ws/outgoing"
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
	Info               ClientInfo
	SkipConnectedCheck bool
	Incoming           Event
}

type ClientInfo struct {
	ID                xid.ID
	RoomID            string
	Authenticated     bool
	AuthenticatedUser string
	Write             chan outgoing.Message
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
		},
		read: read,
	}
	client.debug().Msg("WebSocket New Connection")
	return client
}

// CloseOnError closes the connection.
func (c *Client) CloseOnError(code int, reason string) {
	c.once.Do(func() {
		go func() {
			c.read <- ClientMessage{
				Info: c.info,
				Incoming: &Disconnected{
					Code:   code,
					Reason: reason,
				},
			}
		}()
		c.writeCloseMessage(code, reason)
	})
}

func (c *Client) CloseOnDone(code int, reason string) {
	c.once.Do(func() {
		c.writeCloseMessage(code, reason)
	})
}

func (c *Client) writeCloseMessage(code int, reason string) {
	message := websocket.FormatCloseMessage(code, reason)
	_ = c.conn.WriteControl(websocket.CloseMessage, message, time.Now().Add(writeWait))
	c.conn.Close()
}

// startWriteHandler starts listening on the client connection. As we do not need anything from the client,
// we ignore incoming messages. Leaves the loop on errors.
func (c *Client) startReading(pongWait time.Duration) {
	defer c.CloseOnError(websocket.CloseNormalClosure, "Reader Routine Closed")

	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(appData string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		t, m, err := c.conn.NextReader()
		if err != nil {
			c.CloseOnError(websocket.CloseNormalClosure, "read error: "+err.Error())
			return
		}
		if t == websocket.BinaryMessage {
			c.CloseOnError(websocket.CloseUnsupportedData, "unsupported binary message type")
			return
		}

		incoming, err := ReadTypedIncoming(m)
		if err != nil {
			c.CloseOnError(websocket.CloseUnsupportedData, fmt.Sprintf("malformed message: %s", err))
			return
		}
		c.debug().Interface("event", fmt.Sprintf("%T", incoming)).Interface("payload", incoming).Msg("WebSocket Receive")
		c.read <- ClientMessage{Info: c.info, Incoming: incoming}
	}
}

// startWriteHandler starts the write loop. The method has the following tasks:
// * ping the client in the interval provided as parameter
// * write messages send by the channel to the client
// * on errors exit the loop.
func (c *Client) startWriteHandler(pingPeriod time.Duration) {
	pingTicker := time.NewTicker(pingPeriod)
	defer pingTicker.Stop()
	defer func() {
		c.debug().Msg("WebSocket Done")
	}()
	defer c.conn.Close()
	for {
		select {
		case message := <-c.info.Write:
			if msg, ok := message.(outgoing.CloseWriter); ok {
				c.debug().Str("reason", msg.Reason).Int("code", msg.Code).Msg("WebSocket Close")
				c.CloseOnDone(msg.Code, msg.Reason)
				return
			}

			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			typed, err := ToTypedOutgoing(message)
			c.debug().Interface("event", typed.Type).Interface("payload", typed.Payload).Msg("WebSocket Send")
			if err != nil {
				c.debug().Err(err).Msg("could not get typed message, exiting connection.")
				c.CloseOnError(websocket.CloseNormalClosure, "malformed outgoing "+err.Error())
				continue
			}

			if room, ok := message.(outgoing.Room); ok {
				c.info.RoomID = room.ID
			}

			if err := writeJSON(c.conn, typed); err != nil {
				c.printWebSocketError("write", err)
				c.CloseOnError(websocket.CloseNormalClosure, "write error"+err.Error())
			}
		case <-pingTicker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ping(c.conn); err != nil {
				c.printWebSocketError("ping", err)
				c.CloseOnError(websocket.CloseNormalClosure, "ping timeout")
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
