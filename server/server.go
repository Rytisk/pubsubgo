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

func main() {
	var b = broker.New()
	go b.ProcessMessages()
	go listenForPublishers(b)
	listenForSubscribers(b)
}

func listenForPublishers(brk *broker.Broker) {
	udpConn, _ := net.ListenUDP("udp4", &net.UDPAddr{Port: 8090})

	tlsConf := generateTLSConfig()
	quicConf := &quic.Config{}

	listener, _ := quic.Listen(udpConn, tlsConf, quicConf)

	for {
		conn, _ := listener.Accept(context.Background())
		fmt.Printf("New publisher connection: %s\n", conn.RemoteAddr().String())

		go func(conn quic.Connection) {
			publisher := broker.NewPublisher()
			brk.AddPublisher(publisher)

			stream, _ := conn.AcceptStream(context.Background())

			if _, err := io.Copy(brk, stream); err != nil {
				fmt.Printf("error while reading from publisher: %s\n", err)
			}

			brk.RemovePublisher(publisher)
		}(conn)
	}
}

func listenForSubscribers(brk *broker.Broker) {
	udpConn, _ := net.ListenUDP("udp4", &net.UDPAddr{Port: 8091})

	tlsConf := generateTLSConfig()
	quicConf := &quic.Config{}

	listener, _ := quic.Listen(udpConn, tlsConf, quicConf)

	for {
		conn, _ := listener.Accept(context.Background())
		fmt.Printf("New subscriber connection: %s\n", conn.RemoteAddr().String())

		go func(conn quic.Connection) {
			subscriber := broker.NewSubscriber()
			brk.AddSubscriber(subscriber)

			stream, _ := conn.OpenUniStreamSync(context.Background())

			for msg := range subscriber.Read() {
				if _, err := stream.Write(msg); err != nil {
					fmt.Printf("error while writing to subscriber: %s\n", err)
					break
				}
			}

			brk.RemoveSubscriber(subscriber)
		}(conn)
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
