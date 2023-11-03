package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/quic-go/quic-go"
	"io"
	"math/big"
	"net"
	"pubsubgo/server/broker"
)

type ProcessConnection func(conn *quic.Connection, brk *broker.Broker)

func processSubscriberConnection(conn *quic.Connection, brk *broker.Broker) {
	subscriber := broker.NewClient()
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

func processPublisherConnection(conn *quic.Connection, brk *broker.Broker) {
	publisher := broker.NewClient()
	brk.AddPublisher(publisher)

	stream, _ := (*conn).AcceptStream(context.Background())

	go func(publisher *broker.Client, stream *quic.Stream) {
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

func main() {
	tlsConf := generateTLSConfig()
	brk := broker.New()
	go brk.ProcessMessages()
	go listen(brk, processPublisherConnection, 8090, tlsConf)
	listen(brk, processSubscriberConnection, 8091, tlsConf)
}

func listen(brk *broker.Broker, process ProcessConnection, port int, tlsConf *tls.Config) {
	udpConn, _ := net.ListenUDP("udp4", &net.UDPAddr{Port: port})
	quicConf := &quic.Config{}

	listener, _ := quic.Listen(udpConn, tlsConf, quicConf)

	for {
		conn, _ := listener.Accept(context.Background())
		fmt.Printf("New connection on port '%d': %s\n", port, conn.RemoteAddr().String())

		go process(&conn, brk)
	}
}

func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"pub-sub-go"},
	}
}
