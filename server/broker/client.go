package broker

type Clients struct {
	entries []*Client
	joined  chan *Client
	left    chan *Client
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

type Client struct {
	messages chan []byte
}

func NewClient() *Client {
	return &Client{
		messages: make(chan []byte, 10),
	}
}

func (c *Client) Read() <-chan []byte {
	return c.messages
}
