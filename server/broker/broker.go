package broker

import (
	"log"
)

const SubscriberJoined = "A new subscriber has joined!"
const NoMoreSubscribers = "There are no subscribers listening!"

type PubSubBroker interface {
	AddSubscriber(subscriber *Client)
	RemoveSubscriber(subscriber *Client)
	AddPublisher(publisher *Client)
	RemovePublisher(publisher *Client)
	Write(message []byte) (int, error)
	ProcessMessages(stop <-chan struct{})
}

type Broker struct {
	subscribers      PubSubClients
	publishers       PubSubClients
	subscriberJoined chan *Client
	subscriberLeft   chan *Client
	publisherJoined  chan *Client
	publisherLeft    chan *Client
	messages         chan []byte
}

func New() *Broker {
	return &Broker{
		subscribers: &Clients{
			entries: make([]*Client, 0),
		},
		publishers: &Clients{
			entries: make([]*Client, 0),
		},
		subscriberJoined: make(chan *Client),
		subscriberLeft:   make(chan *Client),
		publisherJoined:  make(chan *Client),
		publisherLeft:    make(chan *Client),
		messages:         make(chan []byte),
	}
}

func (b *Broker) AddSubscriber(subscriber *Client) {
	b.subscriberJoined <- subscriber
}

func (b *Broker) RemoveSubscriber(subscriber *Client) {
	b.subscriberLeft <- subscriber
}

func (b *Broker) AddPublisher(publisher *Client) {
	b.publisherJoined <- publisher
}

func (b *Broker) RemovePublisher(publisher *Client) {
	b.publisherLeft <- publisher
}

func (b *Broker) Write(message []byte) (int, error) {
	b.messages <- message
	return len(message), nil
}

func (b *Broker) ProcessMessages(stop <-chan struct{}) {
	for {
		select {
		case <-stop:
			log.Println("@ Stopping processing")
			//TODO: close and drain channels?
			return

		case msg := <-b.messages:
			log.Println("@ Message received")
			b.subscribers.Fanout(msg)

		case publisher := <-b.publisherJoined:
			log.Println("@ Publisher joined")
			b.publishers.Add(publisher)

			if b.subscribers.IsEmpty() {
				publisher.Write([]byte(NoMoreSubscribers))
			}

		case publisher := <-b.publisherLeft:
			log.Println("@ Publisher left")
			b.publishers.Remove(publisher)
			close(publisher.messages)

		case subscriber := <-b.subscriberJoined:
			log.Println("@ Subscriber joined")
			b.subscribers.Add(subscriber)
			b.publishers.Fanout([]byte(SubscriberJoined))

		case subscriber := <-b.subscriberLeft:
			log.Println("@ Subscriber left")
			b.subscribers.Remove(subscriber)
			close(subscriber.messages)

			if b.subscribers.IsEmpty() {
				b.publishers.Fanout([]byte(NoMoreSubscribers))
			}
		}
	}
}
