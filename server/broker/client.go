package broker

type Clients struct {
	entries []*Client
}

type PubSubClients interface {
	IsEmpty() bool
	Add(client *Client)
	Remove(client *Client)
	Fanout(message []byte)
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
	for _, client := range c.entries {
		client.Write(message)
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

func (c *Client) Write(message []byte) {
	select {
	case c.messages <- message:
	default: // drop messages in case receiver is disconnected or slow
	}
}
