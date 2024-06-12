package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	userStates = make(map[int64]string) // track user states
	userTokens = make(map[int64]string) // store user tokens
	mu         sync.Mutex               // handle concurrent map access
)

func main() {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN must be set")
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			handleMessage(bot, update.Message)
		}
	}
}

func handleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	mu.Lock()
	state := userStates[message.Chat.ID]
	mu.Unlock()

	if message.IsCommand() {
		switch message.Command() {
		case "login":
			handleLoginCommand(bot, message)
		case "search":
			handleSearchCommand(bot, message)
		case "signin":
			handleSigninCommand(bot, message)
		case "start":
			msg := tgbotapi.NewMessage(message.Chat.ID, "Welcome to the XKCD searcher! Please use /login or /signin.")
			bot.Send(msg)
		default:
			msg := tgbotapi.NewMessage(message.Chat.ID, "Unknown command. Please use /login to authenticate, /signin to create a user profile or /search to find comics.")
			bot.Send(msg)
		}
	} else {
		switch state {
		case "awaitingUsername":
			handleUsername(bot, message)
		case "awaitingPassword":
			handlePassword(bot, message)
		case "authenticated":
			handleSearchQuery(bot, message)
		case "awaitingRegUsername":
			handleRegUsername(bot, message)
		case "awaitingRegPassword":
			handleSignin(bot, message)
		default:
			msg := tgbotapi.NewMessage(message.Chat.ID, "Please use /login to authenticate or /search to find comics.")
			bot.Send(msg)
		}
	}
}

func handleLoginCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	mu.Lock()
	userStates[message.Chat.ID] = "awaitingUsername"
	mu.Unlock()
	msg := tgbotapi.NewMessage(message.Chat.ID, "Please enter your username:")
	bot.Send(msg)
}

func handleSigninCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	mu.Lock()
	userStates[message.Chat.ID] = "awaitingRegUsername"
	mu.Unlock()
	msg := tgbotapi.NewMessage(message.Chat.ID, "Please enter your username:")
	bot.Send(msg)
}

func handleSearchCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	mu.Lock()
	state := userStates[message.Chat.ID]
	token := userTokens[message.Chat.ID]
	mu.Unlock()

	if state != "authenticated" || token == "" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "You need to login first. Use /login command.")
		bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "Please enter your search query:")
	bot.Send(msg)
}

func handleUsername(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	mu.Lock()
	userStates[message.Chat.ID] = "awaitingPassword"
	mu.Unlock()
	mu.Lock()
	userTokens[message.Chat.ID] = message.Text
	mu.Unlock()
	msg := tgbotapi.NewMessage(message.Chat.ID, "Please enter your password:")
	bot.Send(msg)
}

func handleRegUsername(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	mu.Lock()
	userStates[message.Chat.ID] = "awaitingRegPassword"
	mu.Unlock()
	mu.Lock()
	userTokens[message.Chat.ID] = message.Text
	mu.Unlock()
	msg := tgbotapi.NewMessage(message.Chat.ID, "Please enter your password:")
	bot.Send(msg)
}

func handlePassword(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	mu.Lock()
	username := userTokens[message.Chat.ID]
	mu.Unlock()
	password := message.Text

	bot.Request(tgbotapi.DeleteMessageConfig{
		ChatID:    message.Chat.ID,
		MessageID: message.MessageID,
	})

	loginData := map[string]string{
		"username": username,
		"password": password,
	}

	jsonData, err := json.Marshal(loginData)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Error encoding JSON"))
		return
	}

	resp, err := http.Post("http://localhost:8080/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Error logging in"))
		return
	}
	defer resp.Body.Close()

	_, errRb := io.ReadAll(resp.Body)
	if errRb != nil {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Error reading response"))
		return
	}

	if resp.StatusCode != http.StatusOK {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Invalid username or password"))
		return
	}

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "token" {
			mu.Lock()
			userStates[message.Chat.ID] = "authenticated"
			userTokens[message.Chat.ID] = cookie.Value
			mu.Unlock()
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "You are now logged in! Use /search to find comics."))
			return
		}
	}

	bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Error retrieving token"))
}

func handleSignin(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	mu.Lock()
	username, exists := userTokens[message.Chat.ID]
	mu.Unlock()
	if !exists {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Please enter your username:"))
		return
	}

	password := message.Text
	role := "user"

	signinData := map[string]string{
		"username": username,
		"password": password,
		"role":     role,
	}
	jsonData, err := json.Marshal(signinData)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Error encoding JSON"))
		return
	}

	resp, err := http.Post("http://localhost:8080/register", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Error during sign in"))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Error reading response"))
		return
	}

	if resp.StatusCode != http.StatusCreated {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Error: %s", string(body))))
		return
	} else { bot.Send(tgbotapi.NewMessage(message.Chat.ID, "You are now signed in! Use /login to enter the system."))}
}

func handleSearchQuery(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	mu.Lock()
	token := userTokens[message.Chat.ID]
	mu.Unlock()

	query := message.Text
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://localhost:8080/pics?search="+url.QueryEscape(query), nil)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Error creating request"))
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Error making request"))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Error reading response"))
		return
	}

	var comicURLs []string
	if err := json.Unmarshal(body, &comicURLs); err != nil {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Error unmarshalling response"))
		return
	}

	if len(comicURLs) > 3 {
		comicURLs = comicURLs[:3]
	}

	var responseText strings.Builder
	for _, url := range comicURLs {
		responseText.WriteString(url + "\n")
	}

	bot.Send(tgbotapi.NewMessage(message.Chat.ID, responseText.String()))

	if string(body) == "null" || resp.StatusCode != http.StatusOK {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "The comic doesn't exist yet, please check back later :("))
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "or try to enter a query in English :)"))
		return
	}
}