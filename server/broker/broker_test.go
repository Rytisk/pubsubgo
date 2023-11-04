package broker

import (
	"strings"
	"testing"
)

func TestWrite(t *testing.T) {
	broker := New()
	broker.subscribers.entries = []*Client{NewClient(), NewClient()}

	stop := make(chan struct{})
	msg := "hi"

	go func() {
		broker.Write([]byte(msg))
	}()

	go func() {
		for _, subscriber := range broker.subscribers.entries {
			actualMsg := <-subscriber.messages
			if string(actualMsg) != msg {
				t.Errorf("Write() got '%s', want '%s'", actualMsg, msg)
			}
		}
		close(stop)
	}()

	broker.ProcessMessages(stop)
}

func TestAddPublisher(t *testing.T) {
	broker := New()
	publisher := NewClient()

	stop := make(chan struct{})

	go func() {
		broker.AddPublisher(publisher)
		close(stop)
	}()

	broker.ProcessMessages(stop)

	if broker.publishers.IsEmpty() {
		t.Errorf("AddPublisher(publisher) did not add a new entry to broker")
	}
}

func TestRemovePublisher(t *testing.T) {
	broker := New()
	publisher := NewClient()

	stop := make(chan struct{})

	go func() {
		broker.AddPublisher(publisher)
		broker.RemovePublisher(publisher)
		close(stop)
	}()

	broker.ProcessMessages(stop)

	if !broker.publishers.IsEmpty() {
		t.Errorf("RemovePublisher(publisher) did not remove an entry from the broker")
	}
}

func TestRemovingNotExistingPublisher(t *testing.T) {
	broker := New()
	publisher := NewClient()

	stop := make(chan struct{})

	go func() {
		broker.RemovePublisher(publisher)
		close(stop)
	}()

	broker.ProcessMessages(stop)
}

func TestAddSubscriber(t *testing.T) {
	broker := New()
	subscriber := NewClient()

	stop := make(chan struct{})

	go func() {
		broker.AddSubscriber(subscriber)
		close(stop)
	}()

	broker.ProcessMessages(stop)

	if broker.subscribers.IsEmpty() {
		t.Errorf("AddSubscriber(subscriber) did not add a new entry to broker")
	}
}

func TestAddingSubscriberInformsPublishers(t *testing.T) {
	broker := New()
	broker.publishers.entries = []*Client{NewClient(), NewClient()}

	subscriber := NewClient()

	stop := make(chan struct{})

	go func() {
		broker.AddSubscriber(subscriber)
	}()

	go func() {
		for _, publisher := range broker.publishers.entries {
			msg := <-publisher.messages
			if !strings.Contains(string(msg), "joined") {
				t.Errorf("got '%s', expected to contain 'joined'", msg)
			}
		}
		close(stop)
	}()

	broker.ProcessMessages(stop)
}

func TestRemoveSubscriber(t *testing.T) {
	broker := New()
	subscriber := NewClient()
	broker.subscribers.entries = []*Client{subscriber}

	stop := make(chan struct{})

	go func() {
		broker.RemoveSubscriber(subscriber)
		close(stop)
	}()

	broker.ProcessMessages(stop)

	if !broker.subscribers.IsEmpty() {
		t.Errorf("RemoveSubscriber(subscriber) did not remove an entry from the broker")
	}
}

func TestRemovingLastSubscriberInformsPublishers(t *testing.T) {
	subscriber := NewClient()
	broker := New()
	broker.publishers.entries = []*Client{NewClient(), NewClient()}
	broker.subscribers.entries = []*Client{subscriber}

	stop := make(chan struct{})

	go func() {
		broker.RemoveSubscriber(subscriber)
	}()

	go func() {
		for _, publisher := range broker.publishers.entries {
			msg := <-publisher.messages
			if !strings.Contains(string(msg), "no subscribers") {
				t.Errorf("got '%s', expected to contain 'no subscribers'", msg)
			}
		}
		close(stop)
	}()

	broker.ProcessMessages(stop)
}

func TestRemovingSubscriberDoesNotInformPublishers(t *testing.T) {
	broker := New()
	publisher := NewClient()
	subscriber := NewClient()
	broker.subscribers.entries = []*Client{subscriber, NewClient()}
	broker.publishers.entries = []*Client{publisher}

	stop := make(chan struct{})

	go func() {
		broker.RemoveSubscriber(subscriber)
		close(publisher.messages)
	}()

	go func() {
		for _, pub := range broker.publishers.entries {
			msg := <-pub.messages
			if strings.Contains(string(msg), "no subscribers") {
				t.Errorf("got '%s', expected no message containing 'no subscribers'", msg)
			}
		}
		close(stop)
	}()

	broker.ProcessMessages(stop)
}
