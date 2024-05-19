package websocket_rw

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"

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
	writeMutex *sync.Mutex
}

func NewWebsocketRW(con *websocket.Conn) *WebsocketRW {
	return &WebsocketRW{
		con:        con,
		cur_reader: nil,
		isEOF:      false,
		readBuff:   make([]byte, 0, buff_size),
		curWriter:  nil,
		writeMutex: &sync.Mutex{},
	}
}

func (rw *WebsocketRW) getNextReader() (io.Reader, error) {
	msg_type, r, err := rw.con.NextReader()
	if err != nil {
		return nil, fmt.Errorf("failed to get next websocket reader: %w", err)
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
		return 0, fmt.Errorf("got invalid json message: %w", err)
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

type ws_msg struct {
	Msg      string `json:"msg"`
	IsStderr bool   `json:"isStderr"`
}

func (rw *WebsocketRW) writeMsg(msg ws_msg, n int) (int, error) {
	rw.writeMutex.Lock()
	defer rw.writeMutex.Unlock()
	if rw.curWriter == nil {
		w, err := rw.con.NextWriter(websocket.TextMessage)
		if err != nil {
			return 0, fmt.Errorf("error getting next websocket writer: %w", err)
		}
		rw.curWriter = w
	}
	err := json.NewEncoder(rw.curWriter).Encode(msg)
	rw.curWriter.Close()
	rw.curWriter = nil
	return n, err
}

type stdoutWriter func([]byte) (int, error)

func (w stdoutWriter) Write(p []byte) (int, error) {
	return w(p)
}

func (rw *WebsocketRW) StdoutWriter() io.Writer {
	return stdoutWriter(func(p []byte) (int, error) {
		return rw.writeMsg(ws_msg{Msg: string(p), IsStderr: false}, len(p))
	})
}

type stderrWriter func([]byte) (int, error)

func (w stderrWriter) Write(p []byte) (int, error) {
	return w(p)
}

func (rw *WebsocketRW) StderrWriter() io.Writer {
	return stderrWriter(func(p []byte) (int, error) {
		return rw.writeMsg(ws_msg{Msg: string(p), IsStderr: true}, len(p))
	})
}

func (rw *WebsocketRW) Close() error {
	if rw.curWriter != nil {
		rw.curWriter.Close()
	}
	return nil
}
