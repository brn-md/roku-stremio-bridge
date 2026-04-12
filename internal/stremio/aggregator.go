package stremio

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Addon represents a Stremio add-on manifest
type Addon struct {
	URL      string   `json:"url"`
	Manifest Manifest `json:"manifest"`
}

type Manifest struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Resources []string `json:"resources"`
	Types     []string `json:"types"`
}

// CatalogResponse optimized for Roku ContentNode
type CatalogItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Poster      string `json:"hdPosterUrl"`
	Type        string `json:"type"`
}

type Aggregator struct {
	Addons []Addon
	Client *http.Client
}

func NewAggregator(addonURLs []string) *Aggregator {
	agg := &Aggregator{
		Client: &http.Client{Timeout: 10 * time.Second},
	}
	for _, u := range addonURLs {
		agg.Addons = append(agg.Addons, Addon{URL: u})
	}
	return agg
}

func (a *Aggregator) FetchCatalog(contentType, id string) ([]CatalogItem, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var results []CatalogItem

	for _, addon := range a.Addons {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			// Stremio catalog URL pattern: {baseUrl}/catalog/{type}/{id}.json
			fullURL := fmt.Sprintf("%s/catalog/%s/%s.json", url, contentType, id)
			
			resp, err := a.Client.Get(fullURL)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			var data struct {
				Metas []CatalogItem `json:"metas"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
				return
			}

			mu.Lock()
			results = append(results, data.Metas...)
			mu.Unlock()
		}(addon.URL)
	}

	wg.Wait()
	return results, nil
}
