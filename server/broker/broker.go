package broker

import "fmt"

type Broker struct {
	subscribers Clients
	publishers  Clients
	messages    chan []byte
}

func New() *Broker {
	return &Broker{
		subscribers: Clients{
			joined:  make(chan *Client),
			left:    make(chan *Client),
			entries: make([]*Client, 0),
		},
		publishers: Clients{
			joined:  make(chan *Client),
			left:    make(chan *Client),
			entries: make([]*Client, 0),
		},
		messages: make(chan []byte),
	}
}

func (b *Broker) AddSubscriber(subscriber *Client) {
	b.subscribers.joined <- subscriber
}

func (b *Broker) RemoveSubscriber(subscriber *Client) {
	b.subscribers.left <- subscriber
}

func (b *Broker) AddPublisher(publisher *Client) {
	b.publishers.joined <- publisher
}

func (b *Broker) RemovePublisher(publisher *Client) {
	b.publishers.left <- publisher
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

		case publisher := <-b.publishers.joined:
			fmt.Println("@ Publisher joined")
			b.publishers.Add(publisher)

		case publisher := <-b.publishers.left:
			fmt.Println("@ Publisher left")
			b.publishers.Remove(publisher)
			close(publisher.messages)

		case subscriber := <-b.subscribers.joined:
			fmt.Println("@ Subscriber joined")
			b.subscribers.Add(subscriber)
			b.publishers.Fanout([]byte("A new subscriber joined!"))

		case subscriber := <-b.subscribers.left:
			fmt.Println("@ Subscriber left")
			b.subscribers.Remove(subscriber)
			close(subscriber.messages)

			if b.subscribers.IsEmpty() {
				b.publishers.Fanout([]byte("There are no subscribers listening!"))
			}
		}
	}
}
