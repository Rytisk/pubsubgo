package broker

type Publisher struct {
	messages chan []byte
}

func NewPublisher() *Publisher {
	return &Publisher{
		messages: make(chan []byte, 10),
	}
}

func (s *Publisher) Read() <-chan []byte {
	return s.messages
}
