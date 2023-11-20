package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os/exec"

	"github.com/kkdai/youtube/v2"
)

type (
	YoutubeSong struct {
		id        string
		title     string
		author    string
		duration  int
		formatURL string
	}

	YoutubeClient struct{}
)

func (y *YoutubeClient) Search(query string) ([]YoutubeSong, error) {
	var finalQuery = query
	_, err := url.ParseRequestURI(query)
	if err != nil {
		finalQuery = fmt.Sprintf("ytsearch:%s", query)
	}
	var out bytes.Buffer
	cmd := exec.Command("yt-dlp", "-vU", "--dump-single-json", "--flat-playlist", finalQuery)
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return nil, err
	}
	str := out.String()
	jMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(str), &jMap)
	if err != nil {
		return nil, err
	}
	var entries []map[string]interface{}
	switch jMap["_type"] {
	case "playlist":
		for _, v := range jMap["entries"].([]interface{}) {
			entries = append(entries, v.(map[string]interface{}))
		}
	case "video":
		entries = append(entries, jMap)
	}
	var videos []YoutubeSong
	for _, v := range entries {
		videos = append(videos, YoutubeSong{
			id:       v["id"].(string),
			title:    v["title"].(string),
			author:   v["uploader"].(string),
			duration: int(v["duration"].(float64)),
		})
	}
	return videos, nil
}

func (v *YoutubeSong) URL() (string, error) {
	if v.formatURL != "" { // cache
		return v.formatURL, nil
	}

	client := youtube.Client{}

	video, err := client.GetVideo(v.id)
	if err != nil {
		panic(err)
	}

	formats := video.Formats.WithAudioChannels() // only get videos with audio
	format := formats.FindByItag(251)            // get the format with itag 251
	if format == nil {
		format = formats.FindByItag(140)
	}
	if format == nil {
		return "", fmt.Errorf("no format found")
	}
	v.formatURL = format.URL
	return format.URL, nil
}

func (v *YoutubeSong) Title() string {
	return v.title
}

func (v *YoutubeSong) Author() string {
	return v.author
}

func (v *YoutubeSong) Duration() string {
	return fmt.Sprintf("%d:%02d", v.duration/60, v.duration%60)
}
