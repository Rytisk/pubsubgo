package broker

import (
	"github.com/stretchr/testify/mock"
	"testing"
)

type ClientsMock struct {
	mock.Mock
}

func (cm *ClientsMock) Fanout(message []byte) {
	cm.Called(message)
}

func (cm *ClientsMock) IsEmpty() bool {
	args := cm.Called()
	return args.Bool(0)
}

func (cm *ClientsMock) Add(client *Client) {
	cm.Called(client)
}

func (cm *ClientsMock) Remove(client *Client) {
	cm.Called(client)
}

func TestWrite(t *testing.T) {
	msg := []byte("hi")

	subscribersMock := &ClientsMock{}
	subscribersMock.On("Fanout", msg).Once()

	broker := New()
	broker.subscribers = subscribersMock

	stop := make(chan struct{})
	go func() {
		broker.Write(msg)
		close(stop)
	}()
	broker.ProcessMessages(stop)

	subscribersMock.AssertExpectations(t)
}

func TestAddPublisher(t *testing.T) {
	publisher := NewClient()

	publishersMock := &ClientsMock{}
	publishersMock.On("Add", publisher).Once()

	broker := New()
	broker.publishers = publishersMock

	stop := make(chan struct{})
	go func() {
		broker.AddPublisher(publisher)
		close(stop)
	}()
	broker.ProcessMessages(stop)

	publishersMock.AssertExpectations(t)
}

func TestRemovePublisher(t *testing.T) {
	publisher := NewClient()

	publishersMock := &ClientsMock{}
	publishersMock.On("Remove", publisher).Once()

	broker := New()
	broker.publishers = publishersMock

	stop := make(chan struct{})
	go func() {
		broker.RemovePublisher(publisher)
		close(stop)
	}()
	broker.ProcessMessages(stop)

	publishersMock.AssertExpectations(t)
}

func TestAddSubscriber(t *testing.T) {
	subscriber := NewClient()
	subscribersMock := &ClientsMock{}
	subscribersMock.On("Add", subscriber).Once()

	broker := New()
	broker.subscribers = subscribersMock

	stop := make(chan struct{})
	go func() {
		broker.AddSubscriber(subscriber)
		close(stop)
	}()
	broker.ProcessMessages(stop)

	subscribersMock.AssertExpectations(t)
}

func TestAddingSubscriberInformsPublishers(t *testing.T) {
	publishersMock := &ClientsMock{}
	publishersMock.On("Fanout", []byte(SubscriberJoined)).Once()

	broker := New()
	broker.publishers = publishersMock

	stop := make(chan struct{})
	go func() {
		broker.AddSubscriber(NewClient())
		close(stop)
	}()
	broker.ProcessMessages(stop)

	publishersMock.AssertExpectations(t)
}

func TestRemoveSubscriber(t *testing.T) {
	subscriber := NewClient()

	clientsMock := &ClientsMock{}
	clientsMock.On("Remove", subscriber).Once()
	clientsMock.On("IsEmpty").Return(false)

	broker := New()
	broker.subscribers = clientsMock

	stop := make(chan struct{})
	go func() {
		broker.RemoveSubscriber(subscriber)
		close(stop)
	}()
	broker.ProcessMessages(stop)

	clientsMock.AssertCalled(t, "Remove", subscriber)
}

func TestRemovingLastSubscriberInformsPublishers(t *testing.T) {
	publishersMock := &ClientsMock{}
	publishersMock.On("Fanout", []byte(NoMoreSubscribers)).Once()

	subscriber := NewClient()

	subscribersMock := &ClientsMock{}
	subscribersMock.On("Remove", subscriber).Once()
	subscribersMock.On("IsEmpty").Return(true)

	broker := New()
	broker.publishers = publishersMock
	broker.subscribers = subscribersMock

	stop := make(chan struct{})
	go func() {
		broker.RemoveSubscriber(subscriber)
		close(stop)
	}()
	broker.ProcessMessages(stop)

	publishersMock.AssertExpectations(t)
	subscribersMock.AssertExpectations(t)
}

func TestRemovingOneSubscriberDoesNotInformPublishers(t *testing.T) {
	publishersMock := &ClientsMock{}
	subscribersMock := &ClientsMock{}
	subscribersMock.On("Remove", mock.Anything).Once()
	subscribersMock.On("IsEmpty").Return(false)

	broker := New()
	subscriber := NewClient()
	broker.subscribers = subscribersMock
	broker.publishers = publishersMock

	stop := make(chan struct{})
	go func() {
		broker.RemoveSubscriber(subscriber)
		close(stop)
	}()
	broker.ProcessMessages(stop)

	publishersMock.AssertNotCalled(t, "Fanout")
}
