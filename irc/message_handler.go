package irc

import (
	"strings"
)

type Message struct {
	From   string
	To     string
	Action string
	Body   string
}

type MessageHandler func(*Message)

type MessageHandlers map[string][]MessageHandler

func createMessage(maskedUser, action, talkedTo, message string) *Message {
	message = strings.TrimPrefix(message, ":")
	maskedUser = strings.TrimPrefix(maskedUser, ":")
	user := strings.SplitN(maskedUser, "!", 2)
	msg := &Message{
		From:      user[0],
		FromIdent: maskedUser,
		To:        talkedTo,
		Action:    action,
		Body:      message,
	}
	return msg
}

func ctcpQuote(cmd string) string {
	quoted := "\001" + cmd + "\001"
	return quoted
}

func ctcpDequote(cmd string) string {
	if cmd[0] != '\001' || cmd[len(cmd)-1] != '\001' {
		return cmd
	}
	return cmd[1 : len(cmd)-1]
}

func (c *Client) setupMsgHandlers() {
	c.msgHandlers = make(MessageHandlers)

	c.AddHandler("PING", func(msg *Message) {
		c.Pong(msg.Body)
	})

	c.AddHandler("PRIVMSG", c.privMsgDefaultHandler)
}

func (c *Client) AddHandler(cmd string, fn MessageHandler) {
	cmd = strings.ToUpper(cmd)
	if _, ok := c.msgHandlers[cmd]; ok {
		c.msgHandlers[cmd] = append(c.msgHandlers[cmd], fn)
	} else {
		c.msgHandlers[cmd] = make([]MessageHandler, 1)
		c.msgHandlers[cmd][0] = fn
	}
}

func (c *Client) privMsgDefaultHandler(msg *Message) {
	if msg.IsCTCP() && c.User.IsMsgForMe(msg) {
		c.handleCTCP(msg)
		return
	}
	//---- handle the usual PRIVMSG here
}

func (c *Client) handleCTCP(msg *Message) {
	cmd := ctcpDequote(msg.Body)
	switch cmd {
	case "VERSION":
		c.responseCTCP(msg.From, "VERSION Sifr:0.0.0")
	case "SOURCE":
		c.responseCTCP(msg.From, "SOURCE https://github.com/fudanchii/sifr")
	}
}

func (m *Message) IsCTCP() bool {
	return m.Body[0] == '\001' && m.Body[len(m.Body)-1] == '\001'
}
