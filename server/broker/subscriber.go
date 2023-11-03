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
