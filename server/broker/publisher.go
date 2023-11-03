package broker

type Publishers struct {
	entries []*Publisher
	joined  chan *Publisher
	left    chan *Publisher
}

func (p *Publishers) Add(publisher *Publisher) {
	p.entries = append(p.entries, publisher)
}

func (p *Publishers) Remove(publisher *Publisher) {
	for idx, other := range p.entries {
		if other == publisher {
			p.entries = append(p.entries[:idx], p.entries[idx+1:]...)
		}
	}
}

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
