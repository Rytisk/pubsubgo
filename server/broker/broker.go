package broker

import "github.com/quic-go/quic-go"

type Broker struct {
	subscriberJoined chan *quic.SendStream
	subscribers      []*quic.SendStream
	messages         chan []byte
}

func New() *Broker {
	return &Broker{
		subscriberJoined: make(chan *quic.SendStream),
		subscribers:      make([]*quic.SendStream, 0),
		messages:         make(chan []byte),
	}
}

func (b *Broker) AddSubscriber(subscriber *quic.SendStream) {
	b.subscriberJoined <- subscriber
}

func (b *Broker) Write(message []byte) (int, error) {
	b.messages <- message
	return len(message), nil
}

func (b *Broker) ProcessMessages() {
	for {
		select {
		case msg := <-b.messages:
			for _, subscriber := range b.subscribers {
				(*subscriber).Write(msg)
			}
		case subscriber := <-b.subscriberJoined:
			b.subscribers = append(b.subscribers, subscriber)
		}
	}
}
