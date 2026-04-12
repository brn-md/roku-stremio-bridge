package debrid

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type RealDebridClient struct {
	APIKey string
	Client *http.Client
}

func NewRealDebridClient(apiKey string) *RealDebridClient {
	return &RealDebridClient{
		APIKey: apiKey,
		Client: &http.Client{Timeout: 15 * time.Second},
	}
}

// ResolveMagnet takes a magnet link and returns a playable stream URL for a specific season/episode
func (rd *RealDebridClient) ResolveMagnet(magnet string, season, episode int) (string, error) {
	// 1. Add Magnet
	torrentID, err := rd.addMagnet(magnet)
	if err != nil {
		return "", fmt.Errorf("failed to add magnet: %w", err)
	}

	// 2. Get Torrent Info to see file list
	info, err := rd.getTorrentInfo(torrentID)
	if err != nil {
		return "", err
	}

	// 3. Find the best file ID matching SxxExx
	fileID := rd.findBestFile(info.Files, season, episode)
	if fileID == "" {
		// Fallback: if no specific file found, select all (might happen for nested movies)
		fileID = "all"
	}

	// 4. Select the specific file
	if err := rd.selectFiles(torrentID, fileID); err != nil {
		return "", fmt.Errorf("failed to select files: %w", err)
	}

	// 5. Get the specific unrestricted link
	// We need to fetch info again after selection to get links
	info, err = rd.getTorrentInfo(torrentID)
	if err != nil || len(info.Links) == 0 {
		return "", fmt.Errorf("failed to get torrent link after selection")
	}

	// 6. Unrestrict the first link (RD creates links for selected files)
	streamURL, err := rd.unrestrictLink(info.Links[0])
	if err != nil {
		return "", fmt.Errorf("failed to unrestrict link: %w", err)
	}

	return streamURL, nil
}

type RDFile struct {
	ID   int    `json:"id"`
	Path string `json:"path"`
}

type TorrentInfo struct {
	Links []string `json:"links"`
	Files []RDFile `json:"files"`
}

func (rd *RealDebridClient) getTorrentInfo(id string) (*TorrentInfo, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://api.real-debrid.com/rest/1.0/torrents/info/%s", id), nil)
	req.Header.Set("Authorization", "Bearer "+rd.APIKey)

	resp, err := rd.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res TorrentInfo
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (rd *RealDebridClient) findBestFile(files []RDFile, season, episode int) string {
	pattern := fmt.Sprintf("s%02de%02d", season, episode) // s01e05
	altPattern := fmt.Sprintf("%dx%02d", season, episode) // 1x05

	for _, f := range files {
		path := strings.ToLower(f.Path)
		if strings.Contains(path, pattern) || strings.Contains(path, altPattern) {
			return fmt.Sprintf("%d", f.ID)
		}
	}
	return ""
}

func (rd *RealDebridClient) addMagnet(magnet string) (string, error) {
	data := url.Values{}
	data.Set("magnet", magnet)

	req, _ := http.NewRequest("POST", "https://api.real-debrid.com/rest/1.0/torrents/addMagnet", strings.NewReader(data.Encode()))
	req.Header.Set("Authorization", "Bearer "+rd.APIKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := rd.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var res struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}
	return res.ID, nil
}

func (rd *RealDebridClient) selectFiles(id, files string) error {
	data := url.Values{}
	data.Set("files", files)

	req, _ := http.NewRequest("POST", fmt.Sprintf("https://api.real-debrid.com/rest/1.0/torrents/selectFiles/%s", id), strings.NewReader(data.Encode()))
	req.Header.Set("Authorization", "Bearer "+rd.APIKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := rd.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// Removed getTorrentLink in favor of integrated getTorrentInfo

func (rd *RealDebridClient) unrestrictLink(link string) (string, error) {
	data := url.Values{}
	data.Set("link", link)

	req, _ := http.NewRequest("POST", "https://api.real-debrid.com/rest/1.0/unrestrict/link", strings.NewReader(data.Encode()))
	req.Header.Set("Authorization", "Bearer "+rd.APIKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := rd.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var res struct {
		Download string `json:"download"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}
	return res.Download, nil
}
