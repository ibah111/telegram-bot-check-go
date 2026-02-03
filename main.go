package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/joho/godotenv"
)

type EnvKey string

type HealthcheckResponce struct {
	Message string `json:"message"`
}

func main() {

	var healthcheck bool = false
	var prev_healthcheck bool = false

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки .env файла")
	}

	var BOT_TOKEN string = os.Getenv("BOT_TOKEN")
	var ADMIN_ID string = os.Getenv("ADMIN_ID")
	var SERVER_URL string = os.Getenv("SERVER_URL")
	var SERVER_PORT string = os.Getenv("SERVER_PORT")
	severUrl := SERVER_URL + ":" + SERVER_PORT

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
	}

	tgbot, err := bot.New(BOT_TOKEN, opts...)
	if err != nil {
		panic(err)
	}

	fmt.Println("Bot started")

	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				fmt.Println("==========TICKER START==========")
				healthcheck = healthcheckRequest(severUrl)

				fmt.Println("Ticker ticked", healthcheck)

				var ok string = "OK"
				if healthcheck == false {
					ok = "NOT OK"
				}

				var Text string = "ibahbalezin.ddns.net " + ok

				fmt.Println(Text)
				fmt.Println("PREV-HEALTHCHECK", prev_healthcheck)
				fmt.Println("HEALTHCHECK", healthcheck)
				if prev_healthcheck != healthcheck {
					tgbot.SendMessage(ctx, &bot.SendMessageParams{
						ChatID: ADMIN_ID,
						Text:   Text,
					})
				}
				prev_healthcheck = healthcheck
				fmt.Println("==========TICKER END============")
			}
		}
	}()

	tgbot.Start(ctx)
}

func handler(ctx context.Context, tgbot *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	var ChatID int64 = update.Message.Chat.ID
	var Text string = update.Message.Text
	fmt.Println("Message received: ", Text, ChatID)
}

func healthcheckRequest(serverUrl string) bool {
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network string, addr string) (net.Conn, error) {
			d := net.Dialer{}
			conn, err := d.DialContext(ctx, network, addr)
			if err != nil {
				return nil, err
			}
			// addr тут — "host:port", но RemoteAddr покажет IP:port
			fmt.Println("Connected to:", conn.RemoteAddr().String())
			return conn, nil
		},
		DisableKeepAlives: true, // чтобы не использовать старое соединение
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	resp, err := client.Get(serverUrl)

	if err != nil {
		fmt.Println("Ошибка при проверке доступности сервера", err)
		return false
	}
	defer resp.Body.Close()
	var responce HealthcheckResponce
	if json_err := json.NewDecoder(resp.Body).Decode(&responce); json_err != nil {
		fmt.Print("***json_err: ", json_err)
		return false
	}
	if responce.Message != "" {
		return true
	}
	return false
}
