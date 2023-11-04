package main

import (
	"context"
	"crypto/tls"
	"github.com/quic-go/quic-go"
	"github.com/stretchr/testify/mock"
	"io"
	"pubsubgo/server/broker"
	"sync"
	"testing"
)

const message = "test"

type WriterMock struct {
	mock.Mock
}

func (wm *WriterMock) Write(p []byte) (n int, err error) {
	args := wm.Called(p)
	return args.Int(0), args.Error(1)
}

// e2e test for the pub-sub server
func TestRunServer(t *testing.T) {
	stop := make(chan struct{})

	messagesCount := 100
	publishersCount := 5
	subscribersCount := 5
	totalMessagesSent := messagesCount * publishersCount

	go run(stop)

	for i := 0; i < publishersCount; i++ {
		if err := runPublisher(t, messagesCount, subscribersCount); err != nil {
			t.Fatalf("Failed to run a publisher, reason: %s", err)
		}
	}

	swg := sync.WaitGroup{}
	wms := make([]*WriterMock, subscribersCount)
	for i := 0; i < subscribersCount; i++ {
		wms[i] = &WriterMock{}
		wms[i].On("Write", mock.Anything).Return(len(message), nil).Times(totalMessagesSent)

		swg.Add(1)
		go runSubscriber(t, wms[i], &swg, totalMessagesSent)
	}
	swg.Wait()
	close(stop)

	for _, wm := range wms {
		wm.AssertExpectations(t)
	}
}

func runSubscriber(t *testing.T, wm *WriterMock, wg *sync.WaitGroup, messagesCount int) {
	defer wg.Done()
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"pub-sub-go"},
	}
	conn, err := quic.DialAddr(context.Background(), "localhost:8091", tlsConf, &quic.Config{})
	if err != nil {
		t.Fatalf("Subscriber failed to dial: %s", err)
	}
	stream, err := conn.AcceptUniStream(context.Background())
	if err != nil {
		t.Fatalf("Subscriber failed to accept stream: %s", err)
	}

	for i := 0; i < messagesCount; i++ {
		buf := make([]byte, len(message))
		if _, err = io.ReadFull(stream, buf); err != nil {
			t.Fatalf("Subscriber failed read a message: %s", err)
		}

		wm.Write(buf)
	}
}

func runPublisher(t *testing.T, messagesCount int, subscribersCount int) error {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"pub-sub-go"},
	}

	conn, err := quic.DialAddr(context.Background(), "localhost:8090", tlsConf, &quic.Config{})
	if err != nil {
		return err
	}
	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		return err
	}
	// Stream is not accepted on the broker until a message is sent over the stream
	if _, err = stream.Write([]byte("Hi from a publisher")); err != nil {
		return err
	}

	go func() {
		// Wait for all subscribers to join before writing the messages
		for i := 0; i < subscribersCount; i++ {
			buf := make([]byte, len(broker.SubscriberJoined))
			if _, err = stream.Read(buf); err != nil {
				t.Fatalf("Publisher failed: %s", err)
			}
		}

		for i := 0; i < messagesCount; i++ {
			if _, err = stream.Write([]byte(message)); err != nil {
				t.Fatalf("Publisher failed: %s", err)
			}
		}
	}()

	return nil
}
