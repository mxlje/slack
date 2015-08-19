package slack

import (
// "fmt"
)

// eventer interface represents a pipe along which messages can be sent
type eventer interface {
	// Write the message to the specified channel
	Write(string, string) error
}

// type Channel string // a Slack channel ID

// Message type represents a message received from Slack
type Message struct {
	// a pipe for sending responses to this message
	// eventStream eventer

	// the strategy for sending the response - depending upon how the message was received e.g. a reply if
	// addressed specifically to the bot or a send if not
	// responseStrategy func(*Message, string) error

	// the text content of the message
	Text string

	// the name of the user the message is from
	From string

	// id of the user the message is from
	fromID string

	// channel on which the message was received
	Channel string
}

// func sendMessageToChannel(msg *Message, channel Channel) {
// 	return msg.eventStream.Write(channel, msg.Text)

// }

// Send a new message on the specified channel
// func (m *Message) Tell(channel string, text string) error {
// 	return m.eventStream.Write(channel, text)
// }

// Send a new message on the channel this message was received on
// func (m *Message) Send(text string) error {
// 	return m.eventStream.Write(m.channel, text)
// }

// Send a reply to the user who sent this message on the same channel it was received on
// func (m *Message) Reply(text string) error {
// 	return m.Send("<@" + m.fromId + ">: " + text)
// }

// Send a message in a way that matches the way in which this message was received e.g.
// if this message was addressed then send a reply back to person who sent the message.
// func (m *Message) Respond(text string) error {
// 	return m.responseStrategy(m, text)
// }

// response strategy for replying
// func reply(m *Message, text string) error {
// 	return m.Reply(text)
// }

// response strategy for sending
// func send(m *Message, text string) error {
// 	return m.Send(text)
// }

// func newMessage(p *Processor, event map[string]interface{}) (*Message, error) {
// 	var (
// 		userFullName string
// 		userID       string
// 	)

// 	user, ok := event["user"]
// 	if ok {
// 		userID = user.(string)
// 		user, exists := p.users[userID]

// 		if exists {
// 			userFullName = user.RealName
// 		}
// 	}

// 	text, ok := event["text"]
// 	if !ok || text == nil {
// 		return nil, fmt.Errorf("No message text found")
// 	}

// 	m := &Message{
// 		eventStream: p,
// 		Text:        event["text"].(string),
// 		From:        userFullName,
// 		fromID:      userID,
// 		channel:     event["channel"].(string),
// 	}

// 	return m, nil
// }
