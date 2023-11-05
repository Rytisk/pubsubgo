package broker

import (
	"context"
	"github.com/quic-go/quic-go"
	"github.com/stretchr/testify/mock"
	"io"
	"testing"
)

type ConnectionMock struct {
	mock.Mock
	quic.Connection
}

type StreamMock struct {
	mock.Mock
	quic.Stream
}

type BrokerMock struct {
	mock.Mock
	PubSubBroker
}

func (cm *ConnectionMock) OpenUniStreamSync(ctx context.Context) (quic.SendStream, error) {
	args := cm.Called(ctx)
	return args.Get(0).(quic.SendStream), args.Error(1)
}

func (cm *ConnectionMock) AcceptStream(ctx context.Context) (quic.Stream, error) {
	args := cm.Called(ctx)
	return args.Get(0).(quic.Stream), args.Error(1)
}

func (sm *StreamMock) Write(b []byte) (int, error) {
	args := sm.Called(b)
	return args.Int(0), args.Error(1)
}

func (sm *StreamMock) Read(p []byte) (n int, err error) {
	args := sm.Called(p)
	p = testMsg
	return args.Int(0), args.Error(1)
}

func (bm *BrokerMock) AddSubscriber(subscriber *Client) {
	bm.Called(subscriber)
	subscriber.messages <- testMsg
}

func (bm *BrokerMock) RemoveSubscriber(subscriber *Client) {
	bm.Called(subscriber)
}

func (bm *BrokerMock) AddPublisher(publisher *Client) {
	bm.Called(publisher)
	publisher.messages <- testMsg
}

func (bm *BrokerMock) RemovePublisher(publisher *Client) {
	bm.Called(publisher)
}

func (bm *BrokerMock) Write(b []byte) (int, error) {
	args := bm.Called(b)
	return args.Int(0), args.Error(1)
}

var testMsg = []byte("test")

func TestForwardMessagesStopsOnChannelClose(t *testing.T) {
	client := NewClient()

	mockStream := &StreamMock{}
	mockStream.On("Write", testMsg).Return(len(testMsg), nil).Once()

	go func() {
		client.messages <- testMsg
		close(client.messages)
	}()

	forwardMessages(mockStream, client)

	mockStream.AssertExpectations(t)
}

func TestForwardMessagesStopsOnEof(t *testing.T) {
	client := NewClient()

	mockStream := &StreamMock{}
	mockStream.On("Write", testMsg).Return(0, io.EOF).Once()

	go func() {
		client.messages <- testMsg
	}()

	forwardMessages(mockStream, client)

	mockStream.AssertExpectations(t)
}

func TestProcessSubscriberConn(t *testing.T) {
	mockStream := &StreamMock{}
	mockStream.On("Write", testMsg).Return(0, io.EOF)

	mockConn := &ConnectionMock{}
	mockConn.On("OpenUniStreamSync", mock.Anything).Return(mockStream, nil).Once()

	mockBroker := &BrokerMock{}
	mockBroker.On("AddSubscriber", mock.Anything).Once()
	mockBroker.On("RemoveSubscriber", mock.Anything).Once()

	ProcessSubscriberConn(mockConn, mockBroker)

	mockConn.AssertExpectations(t)
	mockStream.AssertExpectations(t)
	mockBroker.AssertExpectations(t)
}

func TestProcessPublisherConn(t *testing.T) {
	mockStream := &StreamMock{}
	mockStream.On("Write", mock.Anything).Return(0, io.EOF)
	mockStream.On("Read", mock.Anything).Return(len(testMsg), nil)

	mockConn := &ConnectionMock{}
	mockConn.On("AcceptStream", mock.Anything).Return(mockStream, nil).Once()

	mockBroker := &BrokerMock{}
	mockBroker.On("AddPublisher", mock.Anything).Once()
	mockBroker.On("RemovePublisher", mock.Anything).Once()
	mockBroker.On("Write", mock.Anything).Return(0, io.EOF)

	ProcessPublisherConn(mockConn, mockBroker)

	mockConn.AssertExpectations(t)
	mockStream.AssertExpectations(t)
	mockBroker.AssertExpectations(t)
}
