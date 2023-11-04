package broker

type Clients struct {
	entries []*Client
}

func (c *Clients) IsEmpty() bool {
	return len(c.entries) == 0
}

func (c *Clients) Add(client *Client) {
	c.entries = append(c.entries, client)
}

func (c *Clients) Remove(client *Client) {
	for idx, other := range c.entries {
		if other == client {
			c.entries = append(c.entries[:idx], c.entries[idx+1:]...)
		}
	}
}

func (c *Clients) Fanout(message []byte) {
	for _, subscriber := range c.entries {
		subscriber.messages <- message // TODO: will block if one of the subs disconnected and buffer is full
	}
}

type Client struct {
	messages chan []byte
}

func NewClient() *Client {
	return &Client{
		messages: make(chan []byte, 10),
	}
}

func (c *Client) ReadMessages() <-chan []byte {
	return c.messages
}
