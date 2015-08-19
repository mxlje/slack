package slack

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	// "github.com/james-bowman/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

// Authenticate with Slack to retrieve websocket connection URL.
func handshake(apiUrl string, token string) (*Config, error) {
	resp, err := http.PostForm(apiUrl, url.Values{"token": {token}})

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	var data Config
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Printf("%T\n%s\n%#v\n", err, err, err)

		switch v := err.(type) {
		case *json.SyntaxError:
			log.Println(string(body[v.Offset-40 : v.Offset]))
		}

		log.Printf("%s", body)
		return nil, err
	}

	return &data, nil
}

// Authenticate with Slack and upgrade to websocket connection
func connectAndUpgrade(url string, token string) (*Config, *websocket.Conn, error) {
	config, err := handshake(url, token)

	if err != nil {
		return nil, nil, err
	}

	conn, _, err := websocket.DefaultDialer.Dial(config.Url, http.Header{})

	if err != nil {
		return nil, nil, err
	}

	return config, conn, nil
}

// Connect to Slack using the supplied authentication token
func Connect(token string) (*Connection, error) {
	apiStartUrl := "https://slack.com/api/rtm.start"

	config, conn, err := connectAndUpgrade(apiStartUrl, token)

	if err != nil {
		return nil, err
	}

	c := Connection{
		ws:     conn,
		out:    make(chan []byte, 256),
		in:     make(chan []byte, 256),
		Config: *config,
	}

	c.start(func() (*Config, *websocket.Conn, error) {
		config, con, err := connectAndUpgrade(apiStartUrl, token)
		return config, con, err
	})

	return &c, nil
}
