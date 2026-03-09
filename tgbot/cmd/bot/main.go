package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gotogether/tgbot/internal/apiclient"
	"github.com/gotogether/tgbot/internal/auth"
	"github.com/gotogether/tgbot/internal/bot"
	"github.com/gotogether/tgbot/internal/config"
)

func main() {
	cfg := config.Load()
	if cfg.TelegramBotToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is required")
	}

	log.Printf("Backend URL: %s", cfg.BackendURL)

	api := apiclient.New(cfg.BackendURL)
	authMgr := auth.NewManager(api, cfg.JWTSecret)

	var opts []bot.Option
	if cfg.TelegramAPIURL != "" {
		log.Printf("Telegram API URL override: %s", cfg.TelegramAPIURL)
		opts = append(opts, bot.WithAPIURL(cfg.TelegramAPIURL))
	}

	b, err := bot.New(cfg.TelegramBotToken, api, authMgr, opts...)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-stop
		log.Println("Shutting down bot...")
		b.Stop()
	}()

	b.Start()
}
