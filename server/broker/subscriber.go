package broker

type Subscribers struct {
	entries []*Subscriber
	joined  chan *Subscriber
	left    chan *Subscriber
}

func (s *Subscribers) Add(subscriber *Subscriber) {
	s.entries = append(s.entries, subscriber)
}

func (s *Subscribers) Remove(subscriber *Subscriber) {
	for idx, other := range s.entries {
		if other == subscriber {
			s.entries = append(s.entries[:idx], s.entries[idx+1:]...)
		}
	}
}

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
