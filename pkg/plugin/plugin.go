package plugin

import "github.com/malcolmmadsheep/handshakes-seeker/pkg/queue"

type Response struct {
}

type Request struct {
	Source string `json:"source"`
	Dest   string `json:"dest"`
}

type Plugin interface {
	GetName() string
	DoRequest(Request) (Response, error)
	GetQueueConfig() queue.Config
}
