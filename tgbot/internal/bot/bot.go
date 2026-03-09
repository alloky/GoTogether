package bot

import (
	"context"
	"log"
	"time"

	"github.com/gotogether/tgbot/internal/apiclient"
	"github.com/gotogether/tgbot/internal/auth"
	tele "gopkg.in/telebot.v3"
)

type Option func(*tele.Settings)

func WithAPIURL(url string) Option {
	return func(s *tele.Settings) {
		if url != "" {
			s.URL = url
		}
	}
}

type Bot struct {
	tbot  *tele.Bot
	api   *apiclient.Client
	auth  *auth.Manager
	conv  *ConversationManager
	votes *VoteStore
}

func New(token string, api *apiclient.Client, authMgr *auth.Manager, opts ...Option) (*Bot, error) {
	pref := tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}
	for _, o := range opts {
		o(&pref)
	}

	tb, err := tele.NewBot(pref)
	if err != nil {
		return nil, err
	}

	b := &Bot{
		tbot:  tb,
		api:   api,
		auth:  authMgr,
		conv:  NewConversationManager(),
		votes: NewVoteStore(),
	}

	b.registerHandlers()

	return b, nil
}

func (b *Bot) Start() {
	log.Println("Telegram bot starting...")
	// Set bot commands menu
	_ = b.tbot.SetCommands([]tele.Command{
		{Text: "start", Description: "Welcome & register"},
		{Text: "meetings", Description: "View your meetings"},
		{Text: "calendar", Description: "Upcoming confirmed events"},
		{Text: "new", Description: "Create a new meeting"},
		{Text: "help", Description: "Show help"},
	})
	b.tbot.Start()
}

func (b *Bot) Stop() {
	b.tbot.Stop()
}

func (b *Bot) registerHandlers() {
	b.tbot.Handle("/start", b.handleStart)
	b.tbot.Handle("/meetings", b.handleMeetings)
	b.tbot.Handle("/calendar", b.handleCalendar)
	b.tbot.Handle("/new", b.handleNew)
	b.tbot.Handle("/help", b.handleHelp)
	b.tbot.Handle("/cancel", b.handleCancel)
	b.tbot.Handle("/skip", b.handleSkip)
	b.tbot.Handle("/done", b.handleDone)

	// Callback queries (inline buttons)
	b.tbot.Handle(&tele.Btn{Unique: "view"}, b.cbView)
	b.tbot.Handle(&tele.Btn{Unique: "vote"}, b.cbVote)
	b.tbot.Handle(&tele.Btn{Unique: "vtog"}, b.cbVoteToggle)
	b.tbot.Handle(&tele.Btn{Unique: "vsub"}, b.cbVoteSubmit)
	b.tbot.Handle(&tele.Btn{Unique: "rsvp"}, b.cbRSVP)
	b.tbot.Handle(&tele.Btn{Unique: "conf"}, b.cbConfirm)
	b.tbot.Handle(&tele.Btn{Unique: "cfsl"}, b.cbConfirmSlot)
	b.tbot.Handle(&tele.Btn{Unique: "del"}, b.cbDelete)
	b.tbot.Handle(&tele.Btn{Unique: "mpage"}, b.cbMeetingsPage)
	b.tbot.Handle(&tele.Btn{Unique: "back"}, b.cbBack)
	b.tbot.Handle(&tele.Btn{Unique: "caltag"}, b.cbCalendarTag)
	b.tbot.Handle(&tele.Btn{Unique: "menu"}, b.cbMenu)
	b.tbot.Handle(&tele.Btn{Unique: "vis"}, b.cbVisibility)
	b.tbot.Handle(&tele.Btn{Unique: "addp"}, b.cbAddParticipant)

	// Text handler for conversation flows
	b.tbot.Handle(tele.OnText, b.handleText)
}

// ensureAuth is a helper that authenticates the Telegram user and returns the JWT.
func (b *Bot) ensureAuth(c tele.Context) (string, error) {
	sender := c.Sender()
	return b.auth.EnsureAuth(context.Background(), sender.ID, sender.FirstName, sender.LastName, sender.Username)
}

// ctx returns a background context (telebot.Context is not context.Context).
func ctx() context.Context {
	return context.Background()
}
