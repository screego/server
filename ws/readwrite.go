package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/screego/server/ws/outgoing"
)

type Typed struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

func ToTypedOutgoing(outgoing outgoing.Message) (Typed, error) {
	payload, err := json.Marshal(outgoing)
	if err != nil {
		return Typed{}, err
	}
	return Typed{
		Type:    outgoing.Type(),
		Payload: payload,
	}, nil
}

func ReadTypedIncoming(r io.Reader) (Event, error) {
	typed := Typed{}
	if err := json.NewDecoder(r).Decode(&typed); err != nil {
		return nil, fmt.Errorf("%s e", err)
	}

	create, ok := provider[typed.Type]

	if !ok {
		return nil, errors.New("cannot handle " + typed.Type)
	}

	payload := create()

	if err := json.Unmarshal(typed.Payload, payload); err != nil {
		return nil, fmt.Errorf("incoming payload %s", err)
	}
	return payload, nil
}

var provider = map[string]func() Event{}

func register(t string, incoming func() Event) {
	provider[t] = incoming
}
