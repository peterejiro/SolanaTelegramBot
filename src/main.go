package main

import (
	"context"
	"fmt"
	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/types"
	"github.com/joho/godotenv"
	"github.com/mymmrac/telego"
	"log"
	"os"
	"strings"
)

var solClient *client.Client
var pendingBalanceRequests = map[telego.ChatID]bool{}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	solanaRPCEndpoint := os.Getenv("SOLANA_RPC_ENDPOINT")
	if solanaRPCEndpoint == "" {
		log.Fatal("SOLANA_RPC_ENDPOINT is not set")
	}
	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	if telegramToken == "" {
		log.Fatal("TELEGRAM_TOKEN is not set")
	}
	solClient = client.NewClient(solanaRPCEndpoint)

	bot, err := telego.NewBot(telegramToken, telego.WithDefaultDebugLogger())
	if err != nil {
		log.Fatalf("Failed to create bot: %s", err)
	}

	err = bot.SetMyCommands(&telego.SetMyCommandsParams{
		Commands: []telego.BotCommand{
			{Command: "start", Description: "Start the bot"},
			{Command: "create_account", Description: "Create a new Solana account"},
			{Command: "fund_account", Description: "Fund a Solana account"},
			{Command: "check_balance", Description: "Check Solana account balance"},
			{Command: "list_tokens", Description: "List tokens in a Solana account"},
			{Command: "help", Description: "Show help message"},
		},
	})
	if err != nil {
		log.Fatalf("Failed to set commands: %s", err)
	}

	updates, _ := bot.UpdatesViaLongPolling(nil)
	for update := range updates {
		if update.Message == nil || update.Message.Text == "" {
			continue
		}

		command := update.Message.Text

		log.Printf("Received command: %s", command)

		chatID := update.Message.Chat.ChatID()

		if pendingBalanceRequests[chatID] {
			pendingBalanceRequests[chatID] = false
			handlePublicKeyForBalance(bot, chatID, command)
			continue
		}

		switch command {
		case "/start":
			bot.SendMessage(&telego.SendMessageParams{ChatID: update.Message.Chat.ChatID(), Text: "Welcome to Solana Bot! Use /create_account to create a new account."})
		case "/create_account":
			handleCreateAccount(bot, update.Message.Chat.ChatID())
		case "/fund_account":
			handleFundAccount(bot, update.Message.Chat.ChatID())
		case "/check_balance":
			handleCheckBalance(bot, update.Message.Chat.ChatID())
		case "/list_tokens":
			handleListTokens(bot, update.Message.Chat.ChatID())
		case "/help":
			helpMessage := "Available commands:\n" +
				"/create_account - Create a new Solana account\n" +
				"/fund_account - Fund a Solana account\n" +
				"/check_balance - Check Solana account balance\n" +
				"/list_tokens - List tokens in a Solana account\n" +
				"/help - Show this help message"
			bot.SendMessage(&telego.SendMessageParams{ChatID: update.Message.Chat.ChatID(), Text: helpMessage})
		default:
			_, err := bot.SendMessage(&telego.SendMessageParams{ChatID: update.Message.Chat.ChatID(), Text: "Invalid command"})
			if err != nil {
				return
			}
		}
	}
}

func handleCreateAccount(bot *telego.Bot, chatID telego.ChatID) {
	account := types.NewAccount()
	response := fmt.Sprintf("New Solana account created:\nPublic Key: %s\nPrivate Key: %x", account.PublicKey.ToBase58(), account.PrivateKey)
	bot.SendMessage(&telego.SendMessageParams{ChatID: chatID, Text: response})
}

func handleFundAccount(bot *telego.Bot, chatID telego.ChatID) {
	message := "To fund your Solana account, send some SOL to a valid account address. Use /check_balance to verify the balance."
	bot.SendMessage(&telego.SendMessageParams{ChatID: chatID, Text: message})
}

func handleCheckBalance(bot *telego.Bot, chatID telego.ChatID) {
	bot.SendMessage(&telego.SendMessageParams{ChatID: chatID, Text: "Please reply with the Solana public key to check balance."})
	pendingBalanceRequests[chatID] = true
}

func handleListTokens(bot *telego.Bot, chatID telego.ChatID) {
	bot.SendMessage(&telego.SendMessageParams{ChatID: chatID, Text: "Please reply with the Solana public key to list tokens."})
}

func getBalance(publicKey string) (uint64, error) {
	balance, err := solClient.GetBalance(context.Background(), publicKey)
	if err != nil {
		return 0, err
	}
	return balance, nil
}

func handlePublicKeyForBalance(bot *telego.Bot, chatID telego.ChatID, publicKey string) {
	publicKey = strings.TrimSpace(publicKey)
	if publicKey == "" {
		bot.SendMessage(&telego.SendMessageParams{ChatID: chatID, Text: "Public key cannot be empty. Please try again with a valid key."})
		return
	}

	balance, err := getBalance(publicKey)
	if err != nil {
		bot.SendMessage(&telego.SendMessageParams{ChatID: chatID, Text: fmt.Sprintf("Failed to get balance: %s", err)})
		return
	}

	response := fmt.Sprintf("Balance for account %s: %d lamports", publicKey, balance)
	bot.SendMessage(&telego.SendMessageParams{ChatID: chatID, Text: response})
}
