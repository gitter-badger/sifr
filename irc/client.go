package irc

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type Client struct {
	conn        net.Conn
	user        User
	msgHandlers MessageHandlers
	Errorchan   chan error
}

func Connect(addr string, user User) (*Client, error) {
	client := &Client{
		user:      user,
		Errorchan: make(chan error),
	}
	client.setupMsgHandlers()
	cConn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, &Client{}
	}
	client.conn = cConn
	client.register(user)
	go client.handleInput()
	return client, nil
}

func (c *Client) Error() string {
	return "Error creating client.\n"
}

func (c *Client) Send(cmd string, a ...interface{}) {
	str := fmt.Sprintf(cmd, a...)
	c.conn.Write([]byte(str + "\r\n"))
}

func (c *Client) Join(channel, password string) {
	c.Send("JOIN %s %s", channel, password)
}

func (c *Client) Nick(nick string) {
	c.Send("NICK " + nick)
}

func (c *Client) Notice(to, msg string) {
	c.Send("NOTICE %s :%s", to, msg)
}

func (c *Client) Part(channel string) {
	c.Send("PART " + channel)
}

func (c *Client) Ping(arg string) {
	c.Send("PING :" + arg)
}

func (c *Client) Pong(arg string) {
	c.Send("PONG :" + arg)
}

func (c *Client) PrivMsg(to, msg string) {
	c.Send("PRIVMSG %s :%s", to, msg)
}

func (c *Client) register(user User) {
	c.Nick(user.nick)
	c.Send("USER %s %d * :%s", user.nick, user.mode, user.realname)
}

func (c *Client) responseCTCP(to, answer string) {
	c.Notice(to, ctcpQuote(answer))
}

func (c *Client) respondTo(maskedUser, action, talkedTo, message string) {
	if message[0] == ':' {
		message = message[1:]
	}
	user := strings.SplitN(maskedUser, "!", 2)
	msg := &Message{
		from:   user[0],
		to:     talkedTo,
		action: action,
		body:   message,
	}
	for _, fn := range c.msgHandlers[action] {
		go fn(msg)
	}
}

func (c *Client) handleInput() {
	defer c.conn.Close()
	reader := bufio.NewReader(c.conn)
	for {
		line, err := reader.ReadString('\r')
		if err != nil {
			c.Errorchan <- err
			break
		}
		packet := strings.SplitN(line, " ", 4)
		go c.respondTo(packet[0], packet[1], packet[2], packet[3])
	}

}
