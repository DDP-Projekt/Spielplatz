package websocket_rw

import (
	"encoding/json"
	"errors"
	"io"
	"log"

	"github.com/gorilla/websocket"
)

const buff_size = 128

// implements io.ReadWriter on a websocket connection
type WebsocketRW struct {
	con        *websocket.Conn
	cur_reader io.Reader
	isEOF      bool
	readBuff   []byte
	curWriter  io.WriteCloser
	buffered   int // how many bytes are currently buffered
}

func NewWebsocketRW(con *websocket.Conn) *WebsocketRW {
	return &WebsocketRW{
		con:        con,
		cur_reader: nil,
		isEOF:      false,
		readBuff:   make([]byte, 0, buff_size),
		curWriter:  nil,
		buffered:   0,
	}
}

func (rw *WebsocketRW) getNextReader() (io.Reader, error) {
	msg_type, r, err := rw.con.NextReader()
	if err != nil {
		return nil, err
	}
	if msg_type != websocket.TextMessage {
		return nil, errors.New("expected text message")
	}
	return r, nil
}

func (rw *WebsocketRW) Read(p []byte) (int, error) {
	type Message struct {
		Msg string `json:"msg"`
		Eof bool   `json:"eof"`
	}

	if rw.isEOF {
		return 0, io.EOF
	}

	if rw.cur_reader == nil {
		var err error
		if rw.cur_reader, err = rw.getNextReader(); err != nil {
			log.Printf("getNextReader failed: %s", err)
			rw.isEOF = true
			return 0, err
		}
	}

	if len(rw.readBuff) != 0 {
		n := copy(p, rw.readBuff)
		rw.readBuff = rw.readBuff[n:]
		return n, nil
	}

	var msg Message
	err := json.NewDecoder(rw.cur_reader).Decode(&msg)
	if err != nil {
		log.Printf("error decoding stdin message: %s", err)
		return 0, err
	}

	if msg.Eof {
		rw.isEOF = true
		return 0, io.EOF
	}

	rw.readBuff = []byte(msg.Msg)
	rw.cur_reader = nil
	n := copy(p, rw.readBuff)
	rw.readBuff = rw.readBuff[n:]
	return n, nil
}

func (rw *WebsocketRW) Write(p []byte) (int, error) {
	if rw.curWriter == nil {
		w, err := rw.con.NextWriter(websocket.TextMessage)
		if err != nil {
			return 0, err
		}
		rw.curWriter = w
	}
	n, err := rw.curWriter.Write(p)
	rw.buffered += n
	if rw.buffered >= buff_size {
		rw.curWriter.Close()
		rw.curWriter = nil
		rw.buffered = 0
	}
	return n, err
}

func (rw *WebsocketRW) Close() error {
	if rw.curWriter != nil {
		rw.curWriter.Close()
	}
	return nil
}
