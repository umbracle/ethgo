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

func newWebsocket(url string) (Transport, error) {
	codec, _, err := websocket.DefaultDialer.Dial(url, http.Header{})
	if err != nil {
		return nil, err
	}
	return newStream(codec)
}

// ErrTimeout happens when the websocket requests times out
var ErrTimeout = fmt.Errorf("timeout")

type ackMessage struct {
	buf []byte
	err error
}

type callback func(b []byte, err error)

type stream struct {
	seq   uint64
	codec Codec

	handlerLock sync.Mutex
	handler     map[uint64]callback

	closeCh chan struct{}
	timer   *time.Timer
}

func newStream(codec Codec) (*stream, error) {
	w := &stream{
		codec:   codec,
		closeCh: make(chan struct{}),
		handler: map[uint64]callback{},
	}

	go w.listen()
	return w, nil
}

// Close implements the the transport interface
func (s *stream) Close() error {
	close(s.closeCh)
	return s.codec.Close()
}

func (s *stream) incSeq() uint64 {
	return atomic.AddUint64(&s.seq, 1)
}

func (s *stream) isClosed() bool {
	select {
	case <-s.closeCh:
		return true
	default:
		return false
	}
}

func (s *stream) listen() {
	for {
		var resp codec.Response
		err := s.codec.ReadJSON(&resp)

		if err != nil {
			if !s.isClosed() {
				// log error
			}
			return
		}

		go s.handleMsg(resp)
	}
}

func (s *stream) handleMsg(response codec.Response) {
	s.handlerLock.Lock()
	callback, ok := s.handler[response.ID]
	if !ok {
		s.handlerLock.Unlock()
		return
	}

	// delete handler
	delete(s.handler, response.ID)
	s.handlerLock.Unlock()

	if response.Error != nil {
		callback(nil, response.Error)
	} else {
		callback(response.Result, nil)
	}
}

func (s *stream) setHandler(id uint64, ack chan *ackMessage) {
	callback := func(b []byte, err error) {
		select {
		case ack <- &ackMessage{b, err}:
		default:
		}
	}

	s.handlerLock.Lock()
	s.handler[id] = callback
	s.handlerLock.Unlock()

	s.timer = time.AfterFunc(5*time.Second, func() {
		s.handlerLock.Lock()
		delete(s.handler, id)
		s.handlerLock.Unlock()

		select {
		case ack <- &ackMessage{nil, ErrTimeout}:
		default:
		}
	})
}

// Call implements the transport interface
func (s *stream) Call(method string, out interface{}, params ...interface{}) error {
	seq := s.incSeq()
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

	ack := make(chan *ackMessage)
	s.setHandler(seq, ack)

	if err := s.codec.WriteJSON(request); err != nil {
		return err
	}

	resp := <-ack
	if resp.err != nil {
		return resp.err
	}
	if err := json.Unmarshal(resp.buf, out); err != nil {
		return err
	}
	return nil
}

// Codec is the codec to write and read messages
type Codec interface {
	ReadJSON(v interface{}) error
	WriteJSON(v interface{}) error
	Close() error
}
