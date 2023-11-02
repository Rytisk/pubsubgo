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
)

type loggingWriter struct{}

func (w loggingWriter) Write(b []byte) (int, error) {
	_, err := fmt.Printf("Received message: '%s'\n", string(b))
	return len(b), err
}

func main() {
	go listenForPublishers()

	listenForSubscribers()
}

func listenForPublishers() {
	logger := loggingWriter{}
	udpConn, _ := net.ListenUDP("udp4", &net.UDPAddr{Port: 8090})

	tlsConf := generateTLSConfig()
	quicConf := &quic.Config{}

	listener, _ := quic.Listen(udpConn, tlsConf, quicConf)

	for {
		conn, _ := listener.Accept(context.Background())
		fmt.Printf("New publisher connection: %s\n", conn.RemoteAddr().String())

		go func(conn quic.Connection) {
			stream, _ := conn.AcceptStream(context.Background())
			io.Copy(logger, stream)
		}(conn)
	}
}

func listenForSubscribers() {
	udpConn, _ := net.ListenUDP("udp4", &net.UDPAddr{Port: 8091})

	tlsConf := generateTLSConfig()
	quicConf := &quic.Config{}

	listener, _ := quic.Listen(udpConn, tlsConf, quicConf)

	for {
		conn, _ := listener.Accept(context.Background())
		fmt.Printf("New subscriber connection: %s\n", conn.RemoteAddr().String())

		go func(conn quic.Connection) {
			stream, _ := conn.OpenUniStreamSync(context.Background())
			stream.Write([]byte("hello"))
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
