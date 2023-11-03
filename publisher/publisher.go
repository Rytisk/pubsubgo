package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/quic-go/quic-go"
	"io"
	"os"
	"time"
)

func main() {
	err := publish()

	if err != nil {
		panic(err)
	}
}

type loggingWriter struct{}

func (w loggingWriter) Write(b []byte) (int, error) {
	_, err := fmt.Printf("\n< Received message: '%s'\n> ", string(b))
	return len(b), err
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

	// Stream is not accepted on the broker until a message is sent
	_, err = stream.Write([]byte("A message from a publisher!"))
	if err != nil {
		return err
	}

	go func() {
		io.Copy(loggingWriter{}, stream)
	}()

	for {
		message, err := readUserInput()
		if err != nil {
			return err
		}

		_, err = stream.Write(message)
		if err != nil {
			return err
		}
	}
}

func readUserInput() ([]byte, error) {
	fmt.Print("> ")
	reader := bufio.NewReader(os.Stdin)
	line, _, err := reader.ReadLine()
	return line, err
}
