package broker

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/quic-go/quic-go"
	"io"
	"net"
)

type ProcessConnection func(conn *quic.Connection, brk *Broker)

func ProcessSubscriberConn(conn *quic.Connection, brk *Broker) {
	subscriber := NewClient()
	brk.AddSubscriber(subscriber)

	stream, _ := (*conn).OpenUniStreamSync(context.Background())

	for msg := range subscriber.ReadMessages() {
		if _, err := stream.Write(msg); err != nil {
			fmt.Printf("error while writing to subscriber: %s\n", err)
			break
		}
	}

	brk.RemoveSubscriber(subscriber)
}

func ProcessPublisherConn(conn *quic.Connection, brk *Broker) {
	publisher := NewClient()
	brk.AddPublisher(publisher)

	stream, _ := (*conn).AcceptStream(context.Background())

	go func(publisher *Client, stream *quic.Stream) {
		for msg := range publisher.ReadMessages() {
			if _, err := (*stream).Write(msg); err != nil {
				fmt.Printf("error while writing to publisher: %s\n", err)
				break
			}
		}
	}(publisher, &stream)

	if _, err := io.Copy(brk, stream); err != nil {
		fmt.Printf("error while reading from publisher: %s\n", err)
	}

	brk.RemovePublisher(publisher)
}

func Listen(brk *Broker, process ProcessConnection, port int, tlsConf *tls.Config) {
	udpConn, _ := net.ListenUDP("udp4", &net.UDPAddr{Port: port})
	quicConf := &quic.Config{}

	listener, _ := quic.Listen(udpConn, tlsConf, quicConf)

	for {
		conn, _ := listener.Accept(context.Background())
		fmt.Printf("New connection on port '%d': %s\n", port, conn.RemoteAddr().String())

		go process(&conn, brk)
	}
}
