package beater

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"io/ioutil"

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

	tt := bt.getTrends()

	var err error
	bt.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}

	ticker := time.NewTicker(bt.config.Period)

	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}
		for _, trend := range tt[0].Trends {

			decodedQuery, err := url.QueryUnescape(trend.Query)
			if err != nil {
				log.Fatal(err)
			}

			event := beat.Event{
				Timestamp: tt[0].CreatedAt, //time.Now(),
				Fields: common.MapStr{
					"type": b.Info.Name,
					"name": trend.Name,
					//"url":  decodedURL,
					"query":        decodedQuery,
					"tweet_volume": trend.TweetVolume,
				},
			}
			bt.client.Publish(event)
			logp.Info("Event sent")
		}

	}
}

// Stop stops twitterbeat.
func (bt *Twitterbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}

//
func (bt *Twitterbeat) getTrends() TwitterTrend {
	trendsPlaceURL := "https://api.twitter.com/1.1/trends/place.json?id=23424977"

	client := &http.Client{}

	req, err := http.NewRequest("GET", trendsPlaceURL, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Authorization", bt.config.BearerToken)
	resp, err := client.Do(req)

	body, err := ioutil.ReadAll(resp.Body)

	var jsonBlob = body
	var ttrend TwitterTrend
	err = json.Unmarshal(jsonBlob, &ttrend)
	if err != nil {
		log.Fatal(err)
	}
	return ttrend

}
