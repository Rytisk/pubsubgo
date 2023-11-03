package broker

type Broker struct {
	subscriberJoined chan *Subscriber
	subscriberLeft   chan *Subscriber
	subscribers      []*Subscriber
	publisherJoined  chan *Publisher
	publisherLeft    chan *Publisher
	publishers       []*Publisher
	messages         chan []byte
}

func New() *Broker {
	return &Broker{
		subscriberJoined: make(chan *Subscriber),
		subscriberLeft:   make(chan *Subscriber),
		subscribers:      make([]*Subscriber, 0),
		publisherJoined:  make(chan *Publisher),
		publisherLeft:    make(chan *Publisher),
		publishers:       make([]*Publisher, 0),
		messages:         make(chan []byte),
	}
}

func (b *Broker) AddSubscriber(subscriber *Subscriber) {
	b.subscriberJoined <- subscriber
}

func (b *Broker) RemoveSubscriber(subscriber *Subscriber) {
	b.subscriberLeft <- subscriber
}

func (b *Broker) AddPublisher(publisher *Publisher) {
	b.publisherJoined <- publisher
}

func (b *Broker) RemovePublisher(publisher *Publisher) {
	b.publisherLeft <- publisher
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
				subscriber.messages <- msg // TODO: will block if one of the subs disconnected and buffer is full
			}
		case publisher := <-b.publisherJoined:
			b.publishers = append(b.publishers, publisher)
		case publisher := <-b.publisherLeft:
			for idx, other := range b.publishers {
				if other == publisher {
					b.publishers = append(b.publishers[:idx], b.publishers[idx+1:]...)
				}
			}
			close(publisher.messages)
		case subscriber := <-b.subscriberJoined:
			b.subscribers = append(b.subscribers, subscriber)
		case subscriber := <-b.subscriberLeft:
			for idx, other := range b.subscribers {
				if other == subscriber {
					b.subscribers = append(b.subscribers[:idx], b.subscribers[idx+1:]...)
				}
			}
			close(subscriber.messages)
		}
	}
}
