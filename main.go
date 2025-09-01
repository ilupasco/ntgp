package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
	"time"
)

const (
	urlTelegramAPI = "https://api.telegram.org/bot"
	urlIPApi       = "https://api.ipify.org?format=json"
	defTimeOut     = 1
)

var (
	buildVersion     = "v0.0.0"
	botToken         = "<TOKEN>"
	chatID           = "<CHAT>"
	messagesThreadID = "<TI>"
)

type TelegramMessage struct {
	ParseMode        string `json:"parse_mode"`
	ChatID           string `json:"chat_id"`
	MessagesThreadID string `json:"message_thread_id"`
	Text             string `json:"text"`
}

func NewTelegramMessage(text string) *TelegramMessage {
	return &TelegramMessage{
		ParseMode:        "Markdown",
		ChatID:           chatID,
		MessagesThreadID: messagesThreadID,
		Text:             text,
	}
}

func (tm TelegramMessage) sendMessage() error {
	url := fmt.Sprintf("%s%s/sendMessage", urlTelegramAPI, botToken)
	body, err := json.Marshal(tm)
	if err != nil {
		return err
	}
	ctx, cncl := context.WithTimeout(context.Background(), time.Second*defTimeOut)
	defer cncl()
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return nil
	}
	return nil
}

func getHostName(name *string, wg *sync.WaitGroup) {
	defer wg.Done()
	n, err := os.Hostname()
	if err != nil {
		return
	}
	*name = n
}

func getEnv(name *string, wg *sync.WaitGroup) {
	defer wg.Done()
	user := os.Getenv("USER")
	ip := os.Getenv("SSH_CLIENT")
	ip = strings.Split(ip, " ")[0]
	ssh := "🌐"
	if ip == "" {
		ip = "localhost"
		ssh = "💻"
	}
	*name = fmt.Sprintf("%s %s@%s", ssh, user, ip)
}

func getLocalIP(name *string, wg *sync.WaitGroup) {
	defer wg.Done()
	addrs, err := net.InterfaceAddrs()
	*name = ""
	if err != nil {
		return
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				*name = ipnet.IP.String()
				return
			}
		}
	}
}

func getExternalIP(name *string, wg *sync.WaitGroup) {
	defer wg.Done()
	type Response struct {
		IP string `json:"ip"`
	}
	ctx, cncl := context.WithTimeout(context.Background(), time.Second*defTimeOut)
	defer cncl()
	req, err := http.NewRequestWithContext(ctx, "GET", urlIPApi, nil)
	if err != nil {
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		return
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var response Response
	_ = json.Unmarshal(body, &response)
	*name = response.IP
}

func main() {

	now := time.Now().Format(time.DateTime)
	args := ""
	arg := os.Args[1:]
	if len(arg) > 0 {
		if slices.Contains([]string{"version", "--version", "-V"}, arg[0]) {
			fmt.Printf("%s %s\n", os.Args[0], buildVersion)
			os.Exit(0)
		}
		args = fmt.Sprintf("\n[%s]", strings.Join(arg, " "))
	}

	var wg sync.WaitGroup
	var hostName, envName, localIP, externalIP string

	wg.Add(4)
	go getExternalIP(&externalIP, &wg)
	go getHostName(&hostName, &wg)
	go getEnv(&envName, &wg)
	go getLocalIP(&localIP, &wg)
	wg.Wait()

	text := fmt.Sprintf(
		"*%s* `%s`\n`%s`\n⚡️ `%s` \n⏰ %s%s",
		hostName, localIP, envName, externalIP,
		now,
		args,
	)
	tm := NewTelegramMessage(text)
	_ = tm.sendMessage()
}
