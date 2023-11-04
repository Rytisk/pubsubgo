package broker

import (
	"testing"
)

func TestIsEmpty(t *testing.T) {
	clients := &Clients{
		entries: make([]*Client, 0),
	}

	if !clients.IsEmpty() {
		t.Errorf("IsEmpty() = false")
	}
}

func TestIsNotEmpty(t *testing.T) {
	clients := &Clients{
		entries: make([]*Client, 1),
	}

	if clients.IsEmpty() {
		t.Errorf("IsEmpty() = true")
	}
}

func TestAdd(t *testing.T) {
	clients := &Clients{
		entries: make([]*Client, 0),
	}

	clients.Add(&Client{})

	if clients.IsEmpty() {
		t.Errorf("Add(client) did not add a new client")
	}
}

func TestRemove(t *testing.T) {
	c1 := &Client{}
	c2 := &Client{}
	c3 := &Client{}

	var tests = []struct {
		entries          []*Client
		clientForRemoval *Client
		expectedEntries  []*Client
	}{
		{[]*Client{c1, c2, c3}, c2, []*Client{c1, c3}},
		{[]*Client{c1, c2, c3}, c1, []*Client{c2, c3}},
		{[]*Client{c1, c2, c3}, c3, []*Client{c1, c2}},
		{[]*Client{c1}, c1, []*Client{}},
		{[]*Client{}, c1, []*Client{}},
	}

	for i, test := range tests {
		clients := &Clients{
			entries: test.entries,
		}

		clients.Remove(test.clientForRemoval)

		if !testEq(clients.entries, test.expectedEntries) {
			t.Errorf("test#%d Remove(client) did not remove the correct client entry", i)
		}
	}
}

func TestFanoutEmptyClients(t *testing.T) {
	clients := &Clients{
		entries: []*Client{},
	}

	clients.Fanout([]byte("test"))
}

func TestFanout(t *testing.T) {
	clients := &Clients{
		entries: []*Client{NewClient(), NewClient(), NewClient()},
	}

	msg := []byte("fanout")

	clients.Fanout(msg)

	for i, client := range clients.entries {
		actualMsg := <-client.messages

		if string(actualMsg) != string(msg) {
			t.Errorf("Client #%d got %s, want %s", i, actualMsg, msg)
		}
	}
}

func testEq(a, b []*Client) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
