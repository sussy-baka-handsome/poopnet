package main

import (
	"fmt"
	"net"
	"strings"
)

type Bot struct {
	Conn net.Conn
}

type Command struct {
	Name string
	Args string
}

func NewBot(conn net.Conn) *Bot {
	return &Bot{
		Conn: conn,
	}
}

func NewCommand(command string) (*Command, error) {
	if string(command[0]) != "." {
		return nil, fmt.Errorf("invalid prefix")
	}
	// .shell shut the fuck up
	cmdContent := string(command[1:])
	splitedCmd := strings.Split(cmdContent, " ")
	if len(splitedCmd) < 2 {
		return &Command{
			Name: splitedCmd[0],
			Args: "",
		}, nil
	}
	cmdName := string(splitedCmd[0])
	cmdArgs := string(cmdContent[len(cmdName)+1:])
	return &Command{
		Name: cmdName,
		Args: cmdArgs,
	}, nil
}

func (bot *Bot) Send(data string) {
	bot.Conn.Write([]byte(data))
}

func (bot *Bot) Recv(size int) (string, error) {
	buffer := make([]byte, size)
	if _, err := bot.Conn.Read(buffer); err != nil {
		return "", err
	}
	data := string(buffer)
	data = strings.TrimSpace(data)
	data = strings.Replace(data, "\x00", "", -1)
	return data, nil
}

func (bot *Bot) Handle() {
	defer bot.Conn.Close()
	bot.Send("0x00")
	for {
		data, err := bot.Recv(1024)
		if err != nil {
			break
		}
		if data == "" {
			continue
		}
		if data == "0x01" {
			bot.Send("0x02")
			continue
		}
		command, err := NewCommand(data)
		if err != nil {
			continue
		}
		fmt.Println(command.Name + " " + command.Args)
	}
}

func main() {
	for {
		conn, err := net.Dial("tcp", "127.0.0.1:54321")
		if err != nil {
			break
		}
		bot := NewBot(conn)
		bot.Handle()
	}
}
