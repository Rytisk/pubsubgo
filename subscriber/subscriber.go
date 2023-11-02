package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/quic-go/quic-go"
	"io"
)

type loggingWriter struct{}

func (w loggingWriter) Write(b []byte) (int, error) {
	_, err := fmt.Printf("Received message: '%s'\n", string(b))
	return len(b), err
}

func main() {
	err := subscribe()
	if err != nil {
		panic(err)
	}
}

func subscribe() error {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"pub-sub-go"},
	}

	quicConf := &quic.Config{}

	conn, err := quic.DialAddr(context.Background(), "localhost:8091", tlsConf, quicConf)
	if err != nil {
		return err
	}

	stream, err := conn.AcceptUniStream(context.Background())
	if err != nil {
		return err
	}

	_, err = io.Copy(loggingWriter{}, stream)

	return err
}
