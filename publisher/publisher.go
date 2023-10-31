package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/quic-go/quic-go"
	"time"
)

func main() {
	err := publish()

	if err != nil {
		panic(err)
	}
}

func publish() error {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"pub-sub-go"},
	}

	quicConf := &quic.Config{
		KeepAlivePeriod: 2 * time.Second,
	}

	conn, err := quic.DialAddr(context.Background(), "localhost:8090", tlsConf, quicConf)
	if err != nil {
		return err
	}

	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		return err
	}

	_, err = stream.Write([]byte("message"))
	if err != nil {
		return err
	}

	fmt.Println("Press enter to exit...")
	fmt.Scanln()

	return nil
}
