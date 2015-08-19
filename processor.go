package slack

import (
	"encoding/json"
	"fmt"
	// "regexp"
	"strings"

	log "github.com/Sirupsen/logrus"

	// "github.com/davecgh/go-spew/spew"
)

const (
	eventTypeMessage    = "message"
	eventTypeError      = "error"
	eventTypeHello      = "hello"
	eventTypeUserChange = "user_change"

	maxMessageSize  = 4000
	maxMessageLines = 25
)

type eventPipe chan []byte

type slackEventHandler func(*Processor, map[string]interface{}, []byte)
type slackEventHandlers map[string]slackEventHandler

// Processor type processes inbound events from Slack
type Processor struct {
	// Connection to Slack
	con *Connection

	// Slack user information relating to the user account
	// self User

	// a sequence number to uniquely identify sent messages and correlate with acks from Slack
	sequence int

	// map of event handler functions to handle types of Slack event
	eventHandlers slackEventHandlers

	// map of users who are members of the Slack team
	users map[string]User

	// two channels used for sending and receiving events
	incoming, outgoing eventPipe

	// signal to close the connection
	// shutdown <-chan bool

	state *State
}

// A Gateway holds state and two channels for sending and receiving messages
type Gateway struct {
	Incoming eventPipe
	Outgoing eventPipe
	State    *State
}

// event type represents an event sent TO Slack e.g. messages
type event struct {
	ID      int    `json:"id"`
	Type    string `json:"type"`
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

// user_change event type represents a user change event FROM Slack
type userChangeEvent struct {
	Type        string `json:"type"`
	UpdatedUser User   `json:"user"`
}

// send Event to Slack
func (p *Processor) sendEvent(eventType string, channel string, text string) error {
	p.sequence++

	response := &event{
		ID:      p.sequence,
		Type:    eventType,
		Channel: channel,
		Text:    text,
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		return err
	}

	p.con.Write(responseJSON)

	return nil
}

// Write the message on the specified channel to Slack while respecting the maximum message
// length, splitting up the message if necessary
func (p *Processor) Write(channel string, text string) error {
	for len(text) > 0 {
		lines := strings.Count(text, "\n")
		if len(text) <= maxMessageSize && lines <= maxMessageLines {
			if err := p.sendEvent(eventTypeMessage, channel, text); err != nil {
				return err
			}
			text = ""
		} else {
			// split message at a convenient place
			var breakIndex int
			maxSizeChunk := text

			if len(text) > maxMessageSize {
				maxSizeChunk := text[:maxMessageSize]
				lines = strings.Count(maxSizeChunk, "\n")
			}

			if lines > maxMessageLines {
				var index int
				for n := 0; index < len(maxSizeChunk) && n < maxMessageLines; n++ {
					p := strings.Index(maxSizeChunk[index:], "\n")
					if p == -1 {
						break
					}
					index += p + 1
				}
				breakIndex = index
			} else if lastLineBreak := strings.LastIndex(maxSizeChunk, "\n"); lastLineBreak > -1 {
				breakIndex = lastLineBreak
			} else if lastWordBreak := strings.LastIndexAny(maxSizeChunk, "\n\t .,/\\-(){}[]|=+*&"); lastWordBreak > -1 {
				breakIndex = lastWordBreak
			} else {
				breakIndex = maxMessageSize
			}

			if err := p.sendEvent(eventTypeMessage, channel, text[:breakIndex]); err != nil {
				return err
			}

			if breakIndex != maxMessageSize && lines <= maxMessageLines {
				breakIndex++
			}

			text = text[breakIndex:]
		}
	}

	return nil
}

// Start processing events from Slack. This is the main blocking loop
func (p *Processor) Start() {
	for {
		msg := p.con.Read()

		// log the raw message
		// log.WithField("type", "raw").Printf("%s", msg)

		var data map[string]interface{}
		err := json.Unmarshal(msg, &data)

		if err != nil {
			fmt.Printf("%T\n%s\n%#v\n", err, err, err)
			switch v := err.(type) {
			case *json.SyntaxError:
				fmt.Println(string(msg[v.Offset-40 : v.Offset]))
			}
			log.Printf("%s", msg)
			continue
		}

		// if reply_to attribute is present the event is an ack' for a sent message
		_, isReply := data["reply_to"]
		subtype, ok := data["subtype"]
		var isMessageChangedEvent bool

		if ok {
			isMessageChangedEvent = (subtype.(string) == "message_changed" || subtype.(string) == "message_deleted")
		}

		if !isReply && !isMessageChangedEvent {
			handler, ok := p.eventHandlers[data["type"].(string)]

			if ok {
				handler(p, data, msg)
			}
		}
	}
}

// updateUser updates or adds (if not already existing) the specifed user
func (p *Processor) updateUser(user User) {
	p.state.Users[user.ID] = user
	// log.Println("[INFO] updated user", user.RealName, user.Id)
	// log.Println(p.users)
}

// onConnected is a callback for when the client connects (or reconnects) to Slack.
func (p *Processor) onConnected(con *Connection) {
	log.Printf("Connected to Slack as %s (ID %s)", p.con.Config.Self.Name, p.con.Config.Self.ID)

	for _, user := range con.Config.Users {
		p.updateUser(user)
	}
}

// Starts processing events on the connection from Slack and passes any messages to the hear callback and only
// messages addressed to the bot to the respond callback
func EventProcessor(con *Connection) *Gateway {
	var (
		incoming = make(chan []byte) // Slack -> Server -> Client
		outgoing = make(chan []byte) // Client -> Server -> Slack
	)

	p := Processor{
		con: con,
		// self:     con.Config.Self,
		incoming: incoming,
		outgoing: outgoing,

		eventHandlers: slackEventHandlers{
			eventTypeMessage: func(p *Processor, event map[string]interface{}, rawEvent []byte) {
				// log.Println("MESSAGE", event["text"], p.users[event["user"].(string)].RealName)
				// log.Println(string(rawEvent))
				// log.Printf("%+v", ...).con.config.Users
				// spew.Dump(p.users)
				// spew.Dump(p.con.config.Users)
				// filterMessage(p, event, respond, hear)

				// m, _ := newMessage(p, event)
				// m.eventStream = p

				// spew.Dump(m)

				// for now we simply pipe the raw message to the main app as messages is
				// all we care about
				incoming <- rawEvent
			},

			// The user_change event is sent to all connections for a team when a team
			// member updates their profile or data. Clients can use this to update their local cache of team members.
			eventTypeUserChange: func(p *Processor, event map[string]interface{}, rawEvent []byte) {

				var userEvent userChangeEvent
				err := json.Unmarshal(rawEvent, &userEvent)

				if err != nil {
					fmt.Printf("%T\n%s\n%#v\n", err, err, err)
					switch v := err.(type) {
					case *json.SyntaxError:
						fmt.Println(string(rawEvent[v.Offset-40 : v.Offset]))
					}
					log.Printf("%s", rawEvent)
				}

				log.Infoln("User info changed, ", userEvent.UpdatedUser.Name)
				p.updateUser(userEvent.UpdatedUser)
			},

			// Initial message signaling a successful connection through the realtime API
			eventTypeHello: func(p *Processor, event map[string]interface{}, rawEvent []byte) {
				p.onConnected(con)
			},

			// A generic error from Slack
			eventTypeError: func(p *Processor, event map[string]interface{}, rawEvent []byte) {
				log.Printf("Error received from Slack: %s", rawEvent)
			},
		},
	}

	// Populate state with data from initial connection config
	state := &State{
		Self:  con.Config.Self,
		Users: make(map[string]User),
	}

	// Users
	for _, user := range con.Config.Users {
		// p.updateUser(user)
		state.Users[user.ID] = user
	}

	// Channels
	state.Channels = con.Config.Channels

	// Give state to processor so event handlers can access it
	p.state = state

	gateway := &Gateway{
		Incoming: incoming,
		Outgoing: outgoing,
		State:    state,
	}

	go p.Start()

	return gateway
}

// Invoke one of the specified callbacks for the message if appropriate
// func filterMessage(p *Processor, data map[string]interface{}, respond messageProcessor, hear messageProcessor) {
// 	var userFullName string
// 	var userId string

// 	user, ok := data["user"]
// 	if ok {
// 		userId = user.(string)
// 		user, exists := p.users[userId]
// 		if exists {
// 			userFullName = user.RealName
// 		}
// 	}

// 	// process messages directed at Talbot
// 	r, _ := regexp.Compile("^(<@" + p.self.Id + ">|@?" + p.self.Name + "):? (.+)")

// 	text, ok := data["text"]
// 	if !ok || text == nil {
// 		return
// 	}

// 	matches := r.FindStringSubmatch(text.(string))

// 	if len(matches) == 3 {
// 		if respond != nil {
// m := &Message{eventStream: p, responseStrategy: reply, Text: matches[2], From: userFullName, fromId: userId, channel: data["channel"].(string)}
// 			respond(m)
// 		}
// 	} else if data["channel"].(string)[0] == 'D' {
// 		if respond != nil {
// 			// process direct messages
// 			m := &Message{eventStream: p, responseStrategy: send, Text: text.(string), From: userFullName, fromId: userId, channel: data["channel"].(string)}
// 			respond(m)
// 		}
// 	} else {
// 		if hear != nil {
// 			m := &Message{eventStream: p, responseStrategy: send, Text: text.(string), From: userFullName, fromId: userId, channel: data["channel"].(string)}
// 			hear(m)
// 		}
// 	}
// }
