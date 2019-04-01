package beater

import (
	"fmt"
	"time"
	"net/http"
	"encoding/json"
	"log"
	"os"


	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	"github.com/marcelluseasley/twitterbeat/config"
)

// Twitterbeat configuration.
type Twitterbeat struct {
	done   chan struct{}
	config config.Config
	client beat.Client
}

// TwitterTrend structure contains json structure returned by the Twitter trend API response.
type TwitterTrend []struct {
	Trends []struct {
		Name            string      `json:"name"`
		URL             string      `json:"url"`
		PromotedContent interface{} `json:"promoted_content"`
		Query           string      `json:"query"`
		TweetVolume     int         `json:"tweet_volume"`
	} `json:"trends"`
	AsOf      time.Time `json:"as_of"`
	CreatedAt time.Time `json:"created_at"`
	Locations []struct {
		Name  string `json:"name"`
		Woeid int    `json:"woeid"`
	} `json:"locations"`
}


// New creates an instance of twitterbeat.
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	c := config.DefaultConfig
	if err := cfg.Unpack(&c); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Twitterbeat{
		done:   make(chan struct{}),
		config: c,
	}
	return bt, nil
}

// Run starts twitterbeat.
func (bt *Twitterbeat) Run(b *beat.Beat) error {
	logp.Info("twitterbeat is running! Hit CTRL-C to stop it.")

	var err error
	bt.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}

	ticker := time.NewTicker(bt.config.Period)
	counter := 1
	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}

		event := beat.Event{
			Timestamp: time.Now(),
			Fields: common.MapStr{
				"type":    b.Info.Name,
				"counter": counter,
			},
		}
		bt.client.Publish(event)
		logp.Info("Event sent")
		counter++
	}
}

// Stop stops twitterbeat.
func (bt *Twitterbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}


//
func getTrendingTweet () {
	trendsPlaceURL := "https://api.twitter.com/1.1/trends/place.json?id=23424977"

	client := &http.Client{}

	req, err := http.NewRequest("GET", trendsPlaceURL, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Authorization", os.Getenv(BEARER_TOKEN))
	resp, err := client.Do(req)
}
