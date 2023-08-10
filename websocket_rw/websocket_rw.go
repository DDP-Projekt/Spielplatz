package websocket_rw

import (
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
	cur_writer io.WriteCloser
	buffered   int // how many bytes are currently buffered
}

func NewWebsocketRW(con *websocket.Conn) *WebsocketRW {
	return &WebsocketRW{
		con:        con,
		cur_reader: nil,
		isEOF:      false,
		cur_writer: nil,
		buffered:   0,
	}
}

func (rw *WebsocketRW) Read(p []byte) (int, error) {
	return 0, io.EOF // no input is read
	/*
		if rw.isEOF {
			return 0, io.EOF
		}
		if rw.cur_reader == nil {
			msg_type, r, err := rw.con.NextReader()
			if err != nil {
				rw.isEOF = true
				return 0, io.EOF
			}
			if msg_type != websocket.TextMessage {
				return 0, errors.New("expected text message")
			}
			rw.cur_reader = r
		}
		n, err := rw.cur_reader.Read(p)
		return n, err
	*/
}

func (rw *WebsocketRW) Write(p []byte) (int, error) {
	log.Printf("writing %d bytes\n", len(p))
	if rw.cur_writer == nil {
		w, err := rw.con.NextWriter(websocket.TextMessage)
		if err != nil {
			return 0, err
		}
		rw.cur_writer = w
	}
	n, err := rw.cur_writer.Write(p)
	rw.buffered += n
	if rw.buffered >= buff_size {
		rw.cur_writer.Close()
		rw.cur_writer = nil
		rw.buffered = 0
	}
	return n, err
}

func (rw *WebsocketRW) Close() error {
	if rw.cur_writer != nil {
		rw.cur_writer.Close()
	}
	return nil
}
