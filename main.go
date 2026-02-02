package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

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

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки .env файла")
	}

	var BOT_TOKEN string = os.Getenv("BOT_TOKEN")
	fmt.Println("BOT_TOKEN: ", BOT_TOKEN[:10]+"***")
	var ADMIN_ID string = os.Getenv("ADMIN_ID")
	fmt.Println("ADMIN_ID:", ADMIN_ID)
	var SERVER_URL string = os.Getenv("SERVER_URL")
	fmt.Println("SERVER_URL:", SERVER_URL)
	var SERVER_PORT string = os.Getenv("SERVER_PORT")
	fmt.Println("SERVER_PORT:", SERVER_PORT)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
	}

	b, err := bot.New(BOT_TOKEN, opts...)
	if err != nil {
		panic(err)
	}

	severUrl := SERVER_URL + ":" + SERVER_PORT
	fmt.Println("severUrl:", severUrl)
	resp, err := http.Get(severUrl)
	if err != nil {
		fmt.Println("Ошибка при проверке доступности сервера", err)
		return
	}
	defer resp.Body.Close()

	var responce HealthcheckResponce
	if json_err := json.NewDecoder(resp.Body).Decode(&responce); json_err != nil {
		fmt.Print("***json_err: ", json_err)
		return
	}

	if responce.Message != "" {
		healthcheck = true
	}

	var stringHealcheck string
	if healthcheck == false {
		stringHealcheck = "Not OK"
	} else {
		stringHealcheck = "OK"
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		Text:   "Bot started. Healtcheck: " + stringHealcheck,
		ChatID: ADMIN_ID,
	})
	fmt.Println("Bot started")
	b.Start(ctx)
}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	var ChatID int64 = update.Message.Chat.ID
	var Text string = update.Message.Text
	fmt.Println("Message received: ", Text, ChatID)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: ChatID,
		Text:   Text,
	})
}
