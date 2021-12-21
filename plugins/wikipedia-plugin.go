package plugins

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/malcolmmadsheep/handshakes-seeker/pkg/aconfig"
	"github.com/malcolmmadsheep/handshakes-seeker/pkg/plugin"
	"github.com/malcolmmadsheep/handshakes-seeker/pkg/queue"
)

const WIKIPEDIA_API_BASE_URL = "https://en.wikipedia.org/w/api.php"

type Page struct {
	Title string `json:"title"`
	Links [](struct {
		Title string `json:"title"`
	}) `json:"links"`
}

type WikipediaLinksResponse struct {
	Continue struct {
		Plcontinue string `json:"plcontinue"`
	} `json:"continue"`
	Query struct {
		Pages map[string]json.RawMessage `json:"pages"`
	} `json:"query"`
}

type WikipediaPlugin struct {
}

func (p *WikipediaPlugin) GetName() string {
	return "wikipedia"
}

func (p *WikipediaPlugin) DoRequest(req plugin.Request) (*plugin.Response, error) {
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}

	queryParams := url.Values{
		"action":  {"query"},
		"format":  {"json"},
		"prop":    {"links"},
		"pllimit": {"max"},
		"titles":  {req.SourceUrl},
	}

	if req.Cursor != "" {
		queryParams.Add("plcontinue", req.Cursor)
	}

	apiUrl := fmt.Sprintf("%s?%s", WIKIPEDIA_API_BASE_URL, queryParams.Encode())
	resp, err := client.Get(apiUrl)
	if err != nil {
		return nil, err
	}

	var linksResponse WikipediaLinksResponse

	err = json.NewDecoder(resp.Body).Decode(&linksResponse)
	if err != nil {
		return nil, err
	}

	pages := make([]Page, 0)

	for _, rawPage := range linksResponse.Query.Pages {
		page := Page{}

		err := json.Unmarshal(rawPage, &page)
		if err != nil {
			// add logging
			fmt.Println("ERROR: parsing page", err)
			continue
		}

		pages = append(pages, page)
	}

	pagesCount := 1

	for _, page := range pages {
		pagesCount += len(page.Links)
	}

	pageConnections := make([]plugin.Connection, 0, len(pages)+1)

	for _, page := range pages {
		for _, link := range page.Links {
			pageConnections = append(pageConnections, plugin.Connection{
				SourceUrl: link.Title,
				DestUrl:   req.DestUrl,
			})
		}
	}

	if linksResponse.Continue.Plcontinue != "" {
		pageConnections = append(pageConnections, plugin.Connection{
			SourceUrl: req.SourceUrl,
			DestUrl:   req.DestUrl,
			Cursor:    linksResponse.Continue.Plcontinue,
		})
	}

	response := plugin.Response{
		Connections: pageConnections,
	}

	return &response, nil
}

func (p *WikipediaPlugin) GetQueueConfig() queue.Config {
	delayInMs := aconfig.GetEnvOrInt("HANDSHAKES_WIKI_PLUGIN_DELAY", 500)
	queueSize := aconfig.GetEnvOrInt("HANDSHAKES_WIKI_QUEUE_SIZE", 25)

	return queue.Config{
		Delay:     time.Millisecond * time.Duration(delayInMs),
		QueueSize: uint(queueSize),
	}
}
