package plugin

import "github.com/malcolmmadsheep/handshakes-seeker/pkg/queue"

type Response struct {
}

type Request struct {
	SourceUrl string
	DestUrl   string
	Cursor    string
}

type Plugin interface {
	GetName() string
	DoRequest(Request) (Response, error)
	GetQueueConfig() queue.Config
}
