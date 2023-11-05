# Pub/Sub broker in Go

The broker communicates over QUIC and is using [quic-go](https://github.com/quic-go/quic-go).

The repository includes three main parts:
* [server](https://github.com/Rytisk/pubsubgo/tree/master/server) - the main target of this repository - a Pub/Sub broker that accepts connections using QUIC.
  * The entry point is at `server/server.go`. To run the application use `go run server.go`
  * The server listens on port `8089` for publishers and on `8091` for subscribers.
* [publisher](https://github.com/Rytisk/pubsubgo/tree/master/publisher) - a simple publisher client that allows to send user input to the broker.
* [subscriber](https://github.com/Rytisk/pubsubgo/tree/master/subscriber) - a simple subscriber client that writes all messages received from broker to stdout.
