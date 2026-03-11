package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
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

	// Initialize Sentry (no-op if DSN is empty)
	if cfg.SentryDSN != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:         cfg.SentryDSN,
			Environment: "production",
		}); err != nil {
			log.Printf("WARN: Sentry init failed: %v", err)
		} else {
			log.Println("Sentry initialized")
			defer sentry.Flush(5 * time.Second)
			defer sentry.Recover()
		}
	}

	log.Printf("Backend URL: %s", cfg.BackendURL)

	api := apiclient.New(cfg.BackendURL)
	authMgr := auth.NewManager(api, cfg.JWTSecret, cfg.BotLinkSecret)

	var opts []bot.Option
	if cfg.TelegramAPIURL != "" {
		log.Printf("Telegram API URL override: %s", cfg.TelegramAPIURL)
		opts = append(opts, bot.WithAPIURL(cfg.TelegramAPIURL))
	}

	b, err := bot.New(cfg.TelegramBotToken, api, authMgr, cfg.BotLinkSecret, opts...)
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
