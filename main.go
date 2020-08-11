package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/p1ass/feeder"
	"github.com/pkg/errors"
)

type comic struct {
	Episodes []episode `json:"episodes"`
}

type episode struct {
	ID                    int        `json:"id"`
	Volume                string     `json:"volume"`
	SortVolume            int        `json:"sort_volume"`
	PageCount             int        `json:"page_count"`
	Title                 string     `json:"title"`
	PublishStart          *time.Time `json:"publish_start"`
	PublishEnd            *time.Time `json:"publish_end"`
	MemberPublishStart    *time.Time `json:"member_publish_start"`
	MemberPublishEnd      *time.Time `json:"member_publish_end"`
	Status                string     `json:"status"`
	PageURL               string     `json:"page_url"`
	OgpURL                string     `json:"ogp_url"`
	ListImageURL          string     `json:"list_image_url"`
	ListImageDoubleURL    string     `json:"list_image_double_url"`
	EpisodeNextDate       string     `json:"episode_next_date"`
	NextDateCustomizeText string     `json:"next_date_customize_text"`
	IsUnlimitedComic      bool       `json:"is_unlimited_comic"`
}

type akitaResponse struct {
	Comic comic `json:"comic"`
}

type akitaCrawler struct {
	TitleID string
}

func (crawler *akitaCrawler) Crawl() ([]*feeder.Item, error) {
	url := fmt.Sprintf("https://mangacross.jp/api/comics/%s.json", crawler.TitleID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get response from qiita.")
	}

	var akita akitaResponse
	err = json.NewDecoder(resp.Body).Decode(&akita)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode response body.")
	}

	items := []*feeder.Item{}
	for _, episode := range akita.Comic.Episodes {
		items = append(items, convertEpisodeToItem(episode))
	}
	return items, nil
}

func convertEpisodeToItem(e episode) *feeder.Item {
	return &feeder.Item{
		Title:       e.Title,
		Link:        &feeder.Link{Href: fmt.Sprintf("https://mangacross.jp%s", e.PageURL)},
		Created:     e.PublishStart,
		ID:          fmt.Sprintf("%d", e.ID),
		Description: e.Title,
	}
}

type akitaHandlerBuilder struct {
	TitleID     string
	Title       string
	Link        string
	Description string
	Created     time.Time
}

func (h *akitaHandlerBuilder) buildFeed() (*feeder.Feed, error) {
	crawler := &akitaCrawler{TitleID: h.TitleID}
	items, err := crawler.Crawl()
	if err != nil {
		return &feeder.Feed{}, err
	}

	return &feeder.Feed{
		Title:       h.Title,
		Link:        &feeder.Link{Href: h.Link},
		Description: h.Description,
		Author: &feeder.Author{
			Name:  "",
			Email: ""},
		Created: h.Created,
		Items:   items,
	}, nil
}

func (h *akitaHandlerBuilder) BuildAtomHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		feed, _ := h.buildFeed()
		atomReader, _ := feed.ToAtomReader()
		io.Copy(w, atomReader)
	}
}

func (h *akitaHandlerBuilder) BuildRSSHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		feed, _ := h.buildFeed()
		rssReader, _ := feed.ToRSSReader()
		io.Copy(w, rssReader)
	}
}

func main() {
	yabaiHandlerBuilder := akitaHandlerBuilder{
		TitleID:     "yabai",
		Title:       "僕の心のヤバイやつ",
		Link:        "https://feeds.kuminecraft.xyz",
		Description: "「僕の心のヤバイやつ」の非公式RSSリーダーです",
		Created:     time.Date(2020, time.November, 11, 12, 0, 0, 0, time.UTC),
	}

	http.HandleFunc("/yabai.atom", yabaiHandlerBuilder.BuildAtomHandler())
	http.HandleFunc("/yabai.rss", yabaiHandlerBuilder.BuildRSSHandler())

	// 8080ポートで起動
	http.ListenAndServe(":8080", nil)
}
