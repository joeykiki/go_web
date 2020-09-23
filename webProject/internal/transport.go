package internal

import (
	"encoding/binary"
	"io"
	"log"
	"net"
)

type Transport struct{
	conn net.Conn
}

func NewTransport(conn net.Conn) *Transport{
	return &Transport{conn}
}

func (t *Transport) Send(data []byte){
	buf := make([]byte, 4+len(data))
	binary.BigEndian.PutUint32(buf[:4], uint32(len(data)))
	copy(buf[4:], data)
	_, err := t.conn.Write(buf)
	if err != nil{
		log.Println("send data error ", err)
	}
}


func (t *Transport) Read() ([]byte, error) {
	header := make([]byte, 4)
	_, err := io.ReadFull(t.conn, header)
	if err != nil {
		return nil, err
	}
	dataLen := binary.BigEndian.Uint32(header)
	data := make([]byte, dataLen)
	_, err = io.ReadFull(t.conn, data)
	if err != nil {
		return nil, err
	}
	return data, err
}

