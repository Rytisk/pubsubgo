package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"github.com/quic-go/quic-go"
	"io"
	"math/big"
	"net"
	"os"
)

func main() {
	listen()
}

func listen() {
	udpConn, _ := net.ListenUDP("udp4", &net.UDPAddr{Port: 8090})

	tlsConf := generateTLSConfig()
	quicConf := &quic.Config{}

	listener, _ := quic.Listen(udpConn, tlsConf, quicConf)
	conn, _ := listener.Accept(context.Background())
	stream, _ := conn.AcceptStream(context.Background())

	io.Copy(os.Stdout, stream)
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
