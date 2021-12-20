package plugin

import "github.com/malcolmmadsheep/handshakes-seeker/pkg/queue"

type Connection struct {
	SourceUrl string
	DestUrl   string
	Cursor    string
}

type Response struct {
	Connections []Connection
}

type Request struct {
	SourceUrl string
	DestUrl   string
	Cursor    string
}

type Plugin interface {
	GetName() string
	DoRequest(Request) (*Response, error)
	GetQueueConfig() queue.Config
}
