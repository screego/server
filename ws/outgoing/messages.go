package outgoing

import (
	"encoding/json"

	"github.com/rs/xid"
)

type Message interface {
	Type() string
}

type Room struct {
	ID    string         `json:"id"`
	Mode  ConnectionMode `json:"mode"`
	Users []User         `json:"users"`
}

type User struct {
	ID        xid.ID `json:"id"`
	Name      string `json:"name"`
	Streaming bool   `json:"streaming"`
	You       bool   `json:"you"`
	Owner     bool   `json:"owner"`
}

func (Room) Type() string {
	return "room"
}

type HostSession struct {
	ID         xid.ID      `json:"id"`
	Peer       xid.ID      `json:"peer"`
	ICEServers []ICEServer `json:"iceServers"`
}

func (HostSession) Type() string {
	return "hostsession"
}

type ClientSession struct {
	ID         xid.ID      `json:"id"`
	Peer       xid.ID      `json:"peer"`
	ICEServers []ICEServer `json:"iceServers"`
}

func (ClientSession) Type() string {
	return "clientsession"
}

type ICEServer struct {
	URLs       []string `json:"urls"`
	Credential string   `json:"credential"`
	Username   string   `json:"username"`
}

type P2PMessage struct {
	SID   xid.ID          `json:"sid"`
	Value json.RawMessage `json:"value"`
}

type HostICE P2PMessage

func (HostICE) Type() string {
	return "hostice"
}

type ClientICE P2PMessage

func (ClientICE) Type() string {
	return "clientice"
}

type ClientAnswer P2PMessage

func (ClientAnswer) Type() string {
	return "clientanswer"
}

type HostOffer P2PMessage

func (HostOffer) Type() string {
	return "hostoffer"
}

type EndShare xid.ID

func (EndShare) Type() string {
	return "endshare"
}

type ConnectionMode string

const (
	ConnectionLocal ConnectionMode = "local"
	ConnectionSTUN  ConnectionMode = "stun"
	ConnectionTURN  ConnectionMode = "turn"
)
