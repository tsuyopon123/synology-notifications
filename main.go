package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/caarlos0/env"
	"github.com/nlopes/slack"
)

var appConfig = AppConfig{}
var slackConfig = SlackConfig{}

type SynologyMessage struct {
	Message string `json:"message"`
}

type AppConfig struct {
	ListenPort string `env:"LISTEN_PORT" envDefault:"8080"`
	APIKey     string `env:"API_KEY,required"`
}

type SlackConfig struct {
	Webhook   string `env:"SLACK_WEBHOOK,required"`
	Color     string `env:"SLACK_ATTACHMENT_COLOR" envDefault:"warning"`
	Username  string `env:"SLACK_USERNAME" envDefault:"Synology"`
	IconEmoji string `env:"SLACK_ICONEMOJI" envDefault:":robot_face:"`
}

// PostHandler send notifications from synology to slack
func PostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("api_key") != appConfig.APIKey {
		http.Error(w, "invalid api key", http.StatusUnauthorized)
		log.Printf("invalid api key")
		return
	}

	if r.Method == "POST" {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "error reading request body", http.StatusInternalServerError)
			log.Printf("error reading request body: %s", err)
			return
		}

		synologyMessage := SynologyMessage{}
		err = json.Unmarshal(body, &synologyMessage)
		if err != nil {
			http.Error(w, "error reading request body", http.StatusInternalServerError)
			log.Printf("error reading request body: %s", err)
			return
		}

		msg := slack.WebhookMessage{Username: slackConfig.Username, IconEmoji: slackConfig.IconEmoji, Attachments: []slack.Attachment{{Color: slackConfig.Color, Text: fmt.Sprintf("%s", synologyMessage.Message)}}}

		err = slack.PostWebhook(slackConfig.Webhook, &msg)
		if err != nil {
			http.Error(w, "error sendming slack message", http.StatusInternalServerError)
			log.Printf("error sendming slack message: %s", err)
			return
		}
	} else {
		http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
		return
	}
}

func main() {
	err := env.Parse(&appConfig)
	if err != nil {
		panic(err)
	}

	err = env.Parse(&slackConfig)
	if err != nil {
		panic(err)
	}

	if len(appConfig.APIKey) < 32 {
		panic(fmt.Errorf("api key not long enough it should be 32 characters long not %d", len(appConfig.APIKey)))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", PostHandler)

	log.Printf("listening on port %s", appConfig.ListenPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", appConfig.ListenPort), mux))
}
