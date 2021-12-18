package plugins

import (
	"time"

	"github.com/malcolmmadsheep/handshakes-seeker/pkg/plugin"
	"github.com/malcolmmadsheep/handshakes-seeker/pkg/queue"
)

type WikipediaPlugin struct {
}

func (p *WikipediaPlugin) GetName() string {
	return "wikipedia"
}

func (p *WikipediaPlugin) DoRequest(plugin.Request) (plugin.Response, error) {
	return plugin.Response{}, nil
}

func (p *WikipediaPlugin) GetQueueConfig() queue.Config {
	return queue.Config{
		Delay:     time.Second * 10,
		QueueSize: 25,
	}
}
