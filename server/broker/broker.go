package broker

type Subscriber struct {
	messages chan []byte
}

func NewSubscriber() *Subscriber {
	return &Subscriber{
		messages: make(chan []byte, 10),
	}
}

func (s *Subscriber) Read() <-chan []byte {
	return s.messages
}

type Broker struct {
	subscriberJoined chan *Subscriber
	subscribers      []*Subscriber
	messages         chan []byte
}

func New() *Broker {
	return &Broker{
		subscriberJoined: make(chan *Subscriber),
		subscribers:      make([]*Subscriber, 0),
		messages:         make(chan []byte),
	}
}

func (b *Broker) AddSubscriber(subscriber *Subscriber) {
	b.subscriberJoined <- subscriber
}

func (b *Broker) RemoveSubscriber(subscriber *Subscriber) {
	for idx, other := range b.subscribers {
		if other == subscriber {
			b.subscribers = append(b.subscribers[:idx], b.subscribers[idx+1:]...)
		}
	}
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
		case subscriber := <-b.subscriberJoined:
			b.subscribers = append(b.subscribers, subscriber)
		}
	}
}
