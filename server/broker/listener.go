package broker

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/quic-go/quic-go"
	"io"
	"net"
)

type ProcessConnection func(conn *quic.Connection, brk *Broker)

func ProcessSubscriberConn(conn *quic.Connection, brk *Broker) {
	subscriber := NewClient()
	brk.AddSubscriber(subscriber)
	defer brk.RemoveSubscriber(subscriber)

	stream, err := (*conn).OpenUniStreamSync(context.Background())
	if err != nil {
		fmt.Printf("Failed to open a stream to subscriber, reason: %s\n", err)
		return
	}

	for msg := range subscriber.ReadMessages() {
		if _, err := stream.Write(msg); err != nil {
			if !errors.Is(err, io.EOF) {
				fmt.Printf("error while writing to subscriber: %s\n", err)
			}
			break
		}
	}
}

func ProcessPublisherConn(conn *quic.Connection, brk *Broker) {
	publisher := NewClient()
	brk.AddPublisher(publisher)
	defer brk.RemovePublisher(publisher)

	stream, err := (*conn).AcceptStream(context.Background())
	if err != nil {
		fmt.Printf("Failed to accept publisher's stream, reason: %s\n", err)
		return
	}

	go func(publisher *Client, stream *quic.Stream) {
		for msg := range publisher.ReadMessages() {
			if _, err := (*stream).Write(msg); err != nil {
				if !errors.Is(err, io.EOF) {
					fmt.Printf("error while writing to publisher: %s\n", err)
				}
				break
			}
		}
	}(publisher, &stream)

	if _, err := io.Copy(brk, stream); err != nil {
		if !errors.Is(err, io.EOF) {
			fmt.Printf("error while reading from publisher: %s\n", err)
		}
	}
}

func Listen(brk *Broker, process ProcessConnection, port int, tlsConf *tls.Config) error {
	udpConn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: port})
	if err != nil {
		return err
	}
	quicConf := &quic.Config{}

	listener, err := quic.Listen(udpConn, tlsConf, quicConf)
	if err != nil {
		return err
	}

	for {
		conn, err := listener.Accept(context.Background())
		if err != nil {
			fmt.Printf("Failed to accept a new connection, reason: %s\n", err)
			continue
		}
		fmt.Printf("New connection on port '%d': %s\n", port, conn.RemoteAddr().String())

		go process(&conn, brk)
	}
}
