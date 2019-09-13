package transport

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/umbracle/go-web3/jsonrpc/codec"
)

// ErrTimeout happens when the websocket requests times out
var ErrTimeout = fmt.Errorf("timeout")

type ackMessage struct {
	buf []byte
	err error
}

type callback func(b []byte, err error)

// Websocket is a websocket transport
type Websocket struct {
	seq  uint64
	conn *websocket.Conn

	handlerLock sync.Mutex
	handler     map[uint64]callback

	closeCh chan struct{}
	timer   *time.Timer
}

func newWebsocket(url string) (*Websocket, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, http.Header{})
	if err != nil {
		return nil, err
	}

	w := &Websocket{
		conn:    conn,
		closeCh: make(chan struct{}),
		handler: map[uint64]callback{},
	}

	go w.listen()
	return w, nil
}

// Close implements the the transport interface
func (w *Websocket) Close() error {
	close(w.closeCh)
	return w.conn.Close()
}

func (w *Websocket) incSeq() uint64 {
	return atomic.AddUint64(&w.seq, 1)
}

func (w *Websocket) isClosed() bool {
	select {
	case <-w.closeCh:
		return true
	default:
		return false
	}
}

func (w *Websocket) listen() {
	for {
		typ, buf, err := w.conn.ReadMessage()
		if err != nil {
			if !w.isClosed() {
				// log error
			}
			return
		}
		if typ == websocket.TextMessage {
			go w.handleMsg(buf)
		}
	}
}

func (w *Websocket) handleMsg(buf []byte) {
	var response codec.Response
	if err := json.Unmarshal(buf, &response); err != nil {
		// log failed to decode jsonrpc response
		return
	}

	w.handlerLock.Lock()
	callback, ok := w.handler[response.ID]
	if !ok {
		w.handlerLock.Unlock()
		return
	}

	// delete handler
	delete(w.handler, response.ID)
	w.handlerLock.Unlock()

	if response.Error != nil {
		callback(nil, response.Error)
	} else {
		callback(response.Result, nil)
	}
}

func (w *Websocket) setHandler(id uint64, ack chan *ackMessage) {
	callback := func(b []byte, err error) {
		select {
		case ack <- &ackMessage{b, err}:
		default:
		}
	}

	w.handlerLock.Lock()
	w.handler[id] = callback
	w.handlerLock.Unlock()

	w.timer = time.AfterFunc(5*time.Second, func() {
		w.handlerLock.Lock()
		delete(w.handler, id)
		w.handlerLock.Unlock()

		select {
		case ack <- &ackMessage{nil, ErrTimeout}:
		default:
		}
	})
}

// Call implements the transport interface
func (w *Websocket) Call(method string, out interface{}, params ...interface{}) error {
	seq := w.incSeq()
	request := codec.Request{
		ID:     seq,
		Method: method,
	}
	if len(params) > 0 {
		data, err := json.Marshal(params)
		if err != nil {
			return err
		}
		request.Params = data
	}

	raw, err := json.Marshal(request)
	if err != nil {
		return err
	}

	ack := make(chan *ackMessage)
	w.setHandler(seq, ack)

	if err := w.conn.WriteMessage(websocket.TextMessage, raw); err != nil {
		return err
	}

	resp := <-ack
	if resp.err != nil {
		return err
	}
	if err := json.Unmarshal(resp.buf, out); err != nil {
		return err
	}
	return nil
}
