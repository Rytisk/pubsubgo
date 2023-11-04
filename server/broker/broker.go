package broker

import "fmt"

const SubscriberJoined = "A new subscriber has joined!"
const NoMoreSubscribers = "There are no subscribers listening!"

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
			fmt.Println("@ Stopping processing")
			//TODO: close and drain channels?
			return

		case msg := <-b.messages:
			fmt.Println("@ Message received")
			b.subscribers.Fanout(msg)

		case publisher := <-b.publisherJoined:
			fmt.Println("@ Publisher joined")
			b.publishers.Add(publisher)

		case publisher := <-b.publisherLeft:
			fmt.Println("@ Publisher left")
			b.publishers.Remove(publisher)
			close(publisher.messages)

		case subscriber := <-b.subscriberJoined:
			fmt.Println("@ Subscriber joined")
			b.subscribers.Add(subscriber)
			b.publishers.Fanout([]byte(SubscriberJoined))

		case subscriber := <-b.subscriberLeft:
			fmt.Println("@ Subscriber left")
			b.subscribers.Remove(subscriber)
			close(subscriber.messages)

			if b.subscribers.IsEmpty() {
				b.publishers.Fanout([]byte(NoMoreSubscribers))
			}
		}
	}
}
