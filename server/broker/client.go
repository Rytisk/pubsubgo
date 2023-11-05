package broker

// PubSubClients holds a collection of Client entries and
// provides convenient methods to manage it
type PubSubClients interface {
	IsEmpty() bool
	Add(client *Client)
	Remove(client *Client)
	// Fanout writes the provided message to all the entries in
	// the Clients collection
	Fanout(message []byte)
}

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
	for _, client := range c.entries {
		client.Write(message)
	}
}

// Client is a model depicting a client of the Broker - a Publisher or a Subscriber
type Client struct {
	messages chan []byte
}

func NewClient() *Client {
	return &Client{
		messages: make(chan []byte, 10),
	}
}

// ReadMessages returns a channel of the messages
// sent to this client
func (c *Client) ReadMessages() <-chan []byte {
	return c.messages
}

// Write method writes the provided message to the Client's
// messages channel or drops the message if writing to the channel blocks
func (c *Client) Write(message []byte) {
	select {
	case c.messages <- message:
	default: // drop messages in case receiver is disconnected or slow
	}
}
