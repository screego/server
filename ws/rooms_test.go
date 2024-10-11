package ws

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/xid"
)

const SERVER = "ws://localhost:5050/stream"

func TestMultipleClients(t *testing.T) {
	t.Skip("only for manual testing")
	r := rand.New(rand.NewSource(time.Now().UnixMicro()))

	var wg sync.WaitGroup

	for j := 0; j < 100; j++ {
		name := fmt.Sprint(1)

		users := r.Intn(5000)
		for i := 0; i < users; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				testClient(r.Int63(), name)
			}()
			if i%100 == 0 {
				time.Sleep(10 * time.Millisecond)
			}
		}
		time.Sleep(50 * time.Millisecond)
	}

	wg.Wait()
}

func testClient(i int64, room string) {
	r := rand.New(rand.NewSource(i))
	conn, _, err := websocket.DefaultDialer.Dial(SERVER, nil)
	if err != nil {
		panic(err)
	}
	go func() {
		for {
			_ = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
			_, _, err := conn.ReadMessage()
			if err != nil {
				return
			}
		}
	}()
	defer conn.Close()

	ops := r.Intn(100)
	for i := 0; i < ops; i++ {
		m := msg(r, room)
		err = conn.WriteMessage(websocket.TextMessage, m)
		if err != nil {
			fmt.Println("err", err)
		}
		time.Sleep(30 * time.Millisecond)
	}
}

func msg(r *rand.Rand, room string) []byte {
	typed := Typed{}
	var e Event
	switch r.Intn(8) {
	case 0:
		typed.Type = "clientanswer"
		e = &ClientAnswer{SID: xid.New(), Value: nil}
	case 1:
		typed.Type = "clientice"
		e = &ClientICE{SID: xid.New(), Value: nil}
	case 2:
		typed.Type = "hostice"
		e = &HostICE{SID: xid.New(), Value: nil}
	case 3:
		typed.Type = "hostoffer"
		e = &HostOffer{SID: xid.New(), Value: nil}
	case 4:
		typed.Type = "name"
		e = &Name{UserName: "a"}
	case 5:
		typed.Type = "share"
		e = &StartShare{}
	case 6:
		typed.Type = "stopshare"
		e = &StopShare{}
	case 7:
		typed.Type = "create"
		e = &Create{ID: room, CloseOnOwnerLeave: r.Intn(2) == 0, JoinIfExist: r.Intn(2) == 0, Mode: ConnectionSTUN, UserName: "hello"}
	}

	b, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}
	typed.Payload = json.RawMessage(b)

	b, err = json.Marshal(typed)
	if err != nil {
		panic(err)
	}
	return b
}
