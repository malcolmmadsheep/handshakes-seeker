package plugins

import (
	"fmt"
	"time"

	"github.com/malcolmmadsheep/handshakes-seeker/pkg/plugin"
	"github.com/malcolmmadsheep/handshakes-seeker/pkg/queue"
)

type WikipediaPlugin struct {
}

func (p *WikipediaPlugin) GetName() string {
	return "wikipedia"
}

func (p *WikipediaPlugin) DoRequest(req plugin.Request) (plugin.Response, error) {
	fmt.Println("Wikipedia does request:", req)
	return plugin.Response{}, nil
}

func (p *WikipediaPlugin) GetQueueConfig() queue.Config {
	return queue.Config{
		Delay:     time.Second * 2,
		QueueSize: 25,
	}
}
