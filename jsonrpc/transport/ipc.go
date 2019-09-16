package transport

import (
	"encoding/json"
	"net"
)

func newIPC(addr string) (Transport, error) {
	conn, err := net.Dial("unix", addr)
	if err != nil {
		return nil, err
	}

	codec := &ipcCodec{conn}
	return newStream(codec)
}

func readJSON(c net.Conn, r interface{}) error {
	dec := json.NewDecoder(c)
	if err := dec.Decode(&r); err != nil {
		return err
	}
	return nil
}

type ipcCodec struct {
	c net.Conn
}

func (i *ipcCodec) Close() error {
	return i.c.Close()
}

func (i *ipcCodec) ReadJSON(v interface{}) error {
	dec := json.NewDecoder(i.c)
	if err := dec.Decode(&v); err != nil {
		return err
	}
	return nil
}

func (i *ipcCodec) WriteJSON(v interface{}) error {
	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if _, err := i.c.Write(raw); err != nil {
		return err
	}
	return nil
}
