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

type ProcessConnection func(conn quic.Connection, broker PubSubBroker)

func forwardMessages(dst io.Writer, srcClient *Client) {
	for msg := range srcClient.ReadMessages() {
		if _, err := dst.Write(msg); err != nil {
			if !errors.Is(err, io.EOF) {
				fmt.Printf("error while forwarding: %s\n", err)
			}
			break
		}
	}
}

func ProcessSubscriberConn(conn quic.Connection, broker PubSubBroker) {
	subscriber := NewClient()
	broker.AddSubscriber(subscriber)
	defer broker.RemoveSubscriber(subscriber)

	stream, err := conn.OpenUniStreamSync(context.Background())
	if err != nil {
		fmt.Printf("Failed to open a stream to subscriber, reason: %s\n", err)
		return
	}

	forwardMessages(stream, subscriber)
}

func ProcessPublisherConn(conn quic.Connection, broker PubSubBroker) {
	publisher := NewClient()
	broker.AddPublisher(publisher)
	defer broker.RemovePublisher(publisher)

	stream, err := conn.AcceptStream(context.Background())
	if err != nil {
		fmt.Printf("Failed to accept publisher's stream, reason: %s\n", err)
		return
	}

	go forwardMessages(stream, publisher)

	if _, err := io.Copy(broker, stream); err != nil {
		if !errors.Is(err, io.EOF) {
			fmt.Printf("error while reading from publisher: %s\n", err)
		}
	}
}

func Listen(broker PubSubBroker, process ProcessConnection, port int, tlsConf *tls.Config) error {
	udpConn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: port})
	if err != nil {
		return err
	}

	listener, err := quic.Listen(udpConn, tlsConf, &quic.Config{})
	if err != nil {
		return err
	}

	for {
		conn, err := listener.Accept(context.Background())
		if err != nil {
			fmt.Printf("Failed to accept a new connection, reason: %s\n", err)
			continue
		}

		go process(conn, broker)
	}
}
