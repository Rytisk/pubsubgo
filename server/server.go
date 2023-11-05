package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"pubsubgo/server/broker"
)

func main() {
	stop := make(chan struct{})

	go func(stop chan<- struct{}) {
		fmt.Println("Press enter to stop...")
		fmt.Scanln()
		close(stop)
	}(stop)

	run(stop)
}

func run(stop <-chan struct{}) {
	tlsConf := generateTLSConfig()
	var brk broker.PubSubBroker = broker.New()

	go func() {
		if err := broker.Listen(brk, broker.ProcessPublisherConn, 8090, tlsConf); err != nil {
			fmt.Printf("Failed to start listening for publisher connections, reason: %s\n", err)
		}
	}()

	go func() {
		if err := broker.Listen(brk, broker.ProcessSubscriberConn, 8091, tlsConf); err != nil {
			fmt.Printf("Failed to start listening for subscriber connections, reason: %s\n", err)
		}
	}()

	brk.ProcessMessages(stop)
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
