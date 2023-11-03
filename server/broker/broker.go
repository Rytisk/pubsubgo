package broker

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

func (b *Broker) ProcessMessages() {
	for {
		select {
		case msg := <-b.messages:
			for _, subscriber := range b.subscribers.entries {
				subscriber.messages <- msg // TODO: will block if one of the subs disconnected and buffer is full
			}
		case publisher := <-b.publishers.joined:
			b.publishers.Add(publisher)
		case publisher := <-b.publishers.left:
			b.publishers.Remove(publisher)
			close(publisher.messages)
		case subscriber := <-b.subscribers.joined:
			b.subscribers.Add(subscriber)
		case subscriber := <-b.subscribers.left:
			b.subscribers.Remove(subscriber)
			close(subscriber.messages)
		}
	}
}
