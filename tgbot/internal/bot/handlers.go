package bot

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gotogether/tgbot/internal/apiclient"
	tele "gopkg.in/telebot.v3"
)

const perPage = 5

func (b *Bot) handleStart(c tele.Context) error {
	token, err := b.ensureAuth(c)
	if err != nil {
		log.Printf("Auth error for user %d: %v", c.Sender().ID, err)
		return c.Send("Failed to register. Please try again later.")
	}
	_ = token

	menu := &tele.ReplyMarkup{}
	menu.Inline(
		menu.Row(
			menu.Data("\U0001f4cb My Meetings", "menu", "meetings"),
			menu.Data("\U0001f4c5 Calendar", "menu", "calendar"),
		),
		menu.Row(
			menu.Data("\u2795 New Meeting", "menu", "new"),
			menu.Data("\u2753 Help", "menu", "help"),
		),
	)

	return c.Send(fmt.Sprintf(
		"Welcome to <b>GoTogether</b>, %s! \U0001f389\n\n"+
			"Plan meetings with friends, vote on times, and track your schedule.\n\n"+
			"Use the buttons below or type commands.",
		escapeHTML(c.Sender().FirstName),
	), &tele.SendOptions{ParseMode: tele.ModeHTML}, menu)
}

func (b *Bot) handleMeetings(c tele.Context) error {
	return b.sendMeetingsList(c, 0)
}

func (b *Bot) sendMeetingsList(c tele.Context, page int) error {
	token, err := b.ensureAuth(c)
	if err != nil {
		return c.Send("Please /start first.")
	}

	meetings, err := b.api.ListMyMeetings(ctx(), token)
	if err != nil {
		log.Printf("ListMyMeetings error: %v", err)
		return c.Send("Failed to load meetings. Try again.")
	}

	text := renderMeetingList(meetings, page, perPage)
	markup := b.buildMeetingListKeyboard(meetings, page, perPage)

	return c.Send(text, &tele.SendOptions{ParseMode: tele.ModeHTML}, markup)
}

func (b *Bot) buildMeetingListKeyboard(meetings []apiclient.Meeting, page, pp int) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}
	var rows []tele.Row

	total := len(meetings)
	start := page * pp
	end := start + pp
	if end > total {
		end = total
	}

	for i := start; i < end; i++ {
		m := meetings[i]
		icon := statusEmoji(m.Status)
		rows = append(rows, markup.Row(
			markup.Data(fmt.Sprintf("%s %s", icon, truncate(m.Title, 30)), "view", m.ID),
		))
	}

	// Pagination
	var navBtns []tele.Btn
	if page > 0 {
		navBtns = append(navBtns, markup.Data("\u2b05 Prev", "mpage", fmt.Sprintf("%d", page-1)))
	}
	if end < total {
		navBtns = append(navBtns, markup.Data("Next \u27a1", "mpage", fmt.Sprintf("%d", page+1)))
	}
	if len(navBtns) > 0 {
		rows = append(rows, tele.Row(navBtns))
	}

	rows = append(rows, markup.Row(
		markup.Data("\u2795 New Meeting", "menu", "new"),
	))

	markup.Inline(rows...)
	return markup
}

func (b *Bot) handleCalendar(c tele.Context) error {
	return b.sendCalendar(c, "")
}

func (b *Bot) sendCalendar(c tele.Context, filterTag string) error {
	token, err := b.ensureAuth(c)
	if err != nil {
		return c.Send("Please /start first.")
	}

	meetings, err := b.api.ListAllMeetings(ctx(), token)
	if err != nil {
		log.Printf("ListAllMeetings error: %v", err)
		return c.Send("Failed to load calendar. Try again.")
	}

	// Apply tag filter
	if filterTag != "" {
		var filtered []apiclient.Meeting
		for _, m := range meetings {
			for _, t := range m.Tags {
				if t == filterTag {
					filtered = append(filtered, m)
					break
				}
			}
		}
		meetings = filtered
	}

	text := renderCalendar(meetings)
	if filterTag != "" {
		text = fmt.Sprintf("\U0001f3f7 Filtered by: <b>%s</b>\n\n%s", escapeHTML(filterTag), text)
	}

	// Build tag filter buttons
	markup := b.buildCalendarKeyboard(c, token, filterTag)

	return c.Send(text, &tele.SendOptions{ParseMode: tele.ModeHTML}, markup)
}

func (b *Bot) buildCalendarKeyboard(c tele.Context, token, activeTag string) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}
	var rows []tele.Row

	tags, err := b.api.GetAllTags(ctx(), token)
	if err == nil && len(tags) > 0 {
		var tagBtns []tele.Btn
		if activeTag != "" {
			tagBtns = append(tagBtns, markup.Data("\u274c Clear filter", "caltag", "_clear"))
		}
		for _, t := range tags {
			prefix := ""
			if t == activeTag {
				prefix = "\u2705 "
			}
			tagBtns = append(tagBtns, markup.Data(prefix+truncate(t, 15), "caltag", t))
		}
		// Arrange tag buttons in rows of 3
		for i := 0; i < len(tagBtns); i += 3 {
			end := i + 3
			if end > len(tagBtns) {
				end = len(tagBtns)
			}
			rows = append(rows, tele.Row(tagBtns[i:end]))
		}
	}

	markup.Inline(rows...)
	return markup
}

func (b *Bot) handleLink(c tele.Context) error {
	_, err := b.ensureAuth(c)
	if err != nil {
		return c.Send("Please /start first.")
	}

	data := &ConvData{State: StateAwaitLinkEmail}
	b.conv.Set(c.Chat().ID, data)

	return c.Send(
		"\U0001f517 <b>Link Web Account</b>\n\n"+
			"Enter the <b>email address</b> of your GoTogether web account:\n\n"+
			"A one-time code will be sent to that email.\n"+
			"Send /cancel to abort.",
		&tele.SendOptions{ParseMode: tele.ModeHTML},
	)
}

func (b *Bot) handleNew(c tele.Context) error {
	_, err := b.ensureAuth(c)
	if err != nil {
		return c.Send("Please /start first.")
	}

	data := &ConvData{
		State: StateAwaitTitle,
		Draft: apiclient.CreateMeetingInput{
			IsPublic:          true,
			Tags:              []string{},
			ParticipantEmails: []string{},
			ParticipantIDs:    []string{},
		},
	}
	b.conv.Set(c.Chat().ID, data)

	return c.Send(
		"\u2795 <b>Create New Meeting</b>\n\n"+
			"Step 1/6: What's the <b>title</b> of your meeting?\n\n"+
			"Send /cancel to abort.",
		&tele.SendOptions{ParseMode: tele.ModeHTML},
	)
}

func (b *Bot) handleHelp(c tele.Context) error {
	return c.Send(renderHelp(), &tele.SendOptions{ParseMode: tele.ModeHTML})
}

func (b *Bot) handleCancel(c tele.Context) error {
	b.conv.Clear(c.Chat().ID)
	return c.Send("Cancelled. Use /meetings or /new to continue.")
}

func (b *Bot) handleSkip(c tele.Context) error {
	data := b.conv.Get(c.Chat().ID)
	if data.State == StateIdle {
		return nil
	}
	return b.advanceConversation(c, data, "")
}

func (b *Bot) handleDone(c tele.Context) error {
	data := b.conv.Get(c.Chat().ID)
	if data.State == StateIdle {
		return nil
	}

	switch data.State {
	case StateAwaitTimeSlots, StateAwaitMoreSlots:
		if len(data.Slots) == 0 {
			return c.Send("You need at least one time slot. Please enter one:")
		}
		data.State = StateAwaitTags
		b.conv.Set(c.Chat().ID, data)
		return c.Send(
			"Step 5/6: Add <b>tags</b> (comma-separated) or /skip:\n\n"+
				"Example: <code>work, standup, weekly</code>",
			&tele.SendOptions{ParseMode: tele.ModeHTML},
		)
	case StateAwaitParticipantSearch, StateAwaitMoreParticipants:
		return b.createMeetingFromDraft(c, data)
	default:
		return b.advanceConversation(c, data, "")
	}
}

func (b *Bot) handleText(c tele.Context) error {
	data := b.conv.Get(c.Chat().ID)
	if data.State == StateIdle {
		// Not in conversation — show hint
		return c.Send("Use /meetings, /calendar, /new, or /help")
	}

	return b.advanceConversation(c, data, c.Text())
}

func (b *Bot) advanceConversation(c tele.Context, data *ConvData, input string) error {
	switch data.State {
	case StateAwaitTitle:
		if input == "" {
			return c.Send("Title cannot be empty. Please enter a title:")
		}
		data.Draft.Title = input
		data.State = StateAwaitDescription
		b.conv.Set(c.Chat().ID, data)
		return c.Send(
			"Step 2/6: Enter a <b>description</b> (or /skip):",
			&tele.SendOptions{ParseMode: tele.ModeHTML},
		)

	case StateAwaitDescription:
		data.Draft.Description = input // empty if skipped
		data.State = StateAwaitVisibility
		b.conv.Set(c.Chat().ID, data)

		markup := &tele.ReplyMarkup{}
		markup.Inline(
			markup.Row(
				markup.Data("\U0001f30d Public", "vis", "public"),
				markup.Data("\U0001f512 Private", "vis", "private"),
			),
		)
		return c.Send(
			"Step 3/6: Choose <b>visibility</b>:",
			&tele.SendOptions{ParseMode: tele.ModeHTML},
			markup,
		)

	case StateAwaitTimeSlots, StateAwaitMoreSlots:
		if input == "" {
			return nil // skip/done handled elsewhere
		}
		slot, err := parseTimeSlot(input)
		if err != nil {
			return c.Send(fmt.Sprintf(
				"Invalid format. Use: <code>YYYY-MM-DD HH:MM - HH:MM</code>\n\nError: %s",
				escapeHTML(err.Error()),
			), &tele.SendOptions{ParseMode: tele.ModeHTML})
		}
		data.Slots = append(data.Slots, *slot)
		data.State = StateAwaitMoreSlots
		b.conv.Set(c.Chat().ID, data)
		return c.Send(fmt.Sprintf(
			"\u2705 Time slot %d added.\n\nSend another slot or /done to continue.",
			len(data.Slots),
		))

	case StateAwaitTags:
		if input != "" {
			tags := strings.Split(input, ",")
			for _, t := range tags {
				t = strings.TrimSpace(t)
				if t != "" {
					data.Draft.Tags = append(data.Draft.Tags, t)
				}
			}
		}
		data.State = StateAwaitParticipantSearch
		b.conv.Set(c.Chat().ID, data)
		return c.Send(
			"Step 6/6: <b>Invite participants</b>\n\n"+
				"Type a name to search users, or /done to skip.",
			&tele.SendOptions{ParseMode: tele.ModeHTML},
		)

	case StateAwaitParticipantSearch, StateAwaitMoreParticipants:
		if input == "" {
			return nil
		}
		// Search for users
		token, err := b.ensureAuth(c)
		if err != nil {
			return c.Send("Auth error. /start again.")
		}
		users, err := b.api.SearchUsers(ctx(), token, input)
		if err != nil {
			return c.Send("Search failed. Try again or /done to finish.")
		}
		if len(users) == 0 {
			return c.Send("No users found. Try another name or /done to finish.")
		}

		markup := &tele.ReplyMarkup{}
		var rows []tele.Row
		for _, u := range users {
			rows = append(rows, markup.Row(
				markup.Data(
					fmt.Sprintf("\u2795 %s", u.DisplayName),
					"addp", u.ID,
				),
			))
		}
		rows = append(rows, markup.Row(markup.Data("\u2705 Done adding", "addp", "_done")))
		markup.Inline(rows...)

		return c.Send(renderSearchResults(users), &tele.SendOptions{ParseMode: tele.ModeHTML}, markup)

	case StateAwaitLinkEmail:
		if input == "" {
			return c.Send("Please enter your web account email:")
		}
		// Basic email validation
		if !strings.Contains(input, "@") || !strings.Contains(input, ".") {
			return c.Send("That doesn't look like a valid email. Please try again:")
		}
		email := strings.TrimSpace(input)

		err := b.api.InitiateLinkFromBot(ctx(), b.botLinkSecret, c.Sender().ID, email)
		if err != nil {
			log.Printf("InitiateLinkFromBot error: %v", err)
			b.conv.Clear(c.Chat().ID)
			return c.Send(fmt.Sprintf("Failed to send code: %s\n\nUse /link to try again.", err.Error()))
		}

		data.LinkEmail = email
		data.State = StateAwaitLinkCode
		b.conv.Set(c.Chat().ID, data)
		return c.Send(
			fmt.Sprintf("\u2709 A 6-digit code has been sent to <b>%s</b>.\n\n"+
				"Enter the code here:\n\n"+
				"Send /cancel to abort.", escapeHTML(email)),
			&tele.SendOptions{ParseMode: tele.ModeHTML},
		)

	case StateAwaitLinkCode:
		if input == "" {
			return c.Send("Please enter the 6-digit code:")
		}
		code := strings.TrimSpace(input)
		if len(code) != 6 {
			return c.Send("The code should be 6 digits. Please try again:")
		}

		token, err := b.api.ConfirmLinkFromBot(ctx(), b.botLinkSecret, c.Sender().ID, data.LinkEmail, code)
		if err != nil {
			log.Printf("ConfirmLinkFromBot error: %v", err)
			return c.Send(fmt.Sprintf("Invalid or expired code. Please try again or /cancel:\n\n%s", err.Error()))
		}

		// Update cached JWT to the web user's token
		b.auth.SetToken(c.Sender().ID, token)
		b.conv.Clear(c.Chat().ID)

		return c.Send(
			fmt.Sprintf("\u2705 <b>Account linked!</b>\n\n"+
				"Your Telegram account is now connected to <b>%s</b>.\n"+
				"All your meetings and votes have been merged.",
				escapeHTML(data.LinkEmail)),
			&tele.SendOptions{ParseMode: tele.ModeHTML},
		)

	default:
		b.conv.Clear(c.Chat().ID)
		return c.Send("Something went wrong. Use /new to start over.")
	}
}

func (b *Bot) createMeetingFromDraft(c tele.Context, data *ConvData) error {
	token, err := b.ensureAuth(c)
	if err != nil {
		return c.Send("Auth error. /start again.")
	}

	data.Draft.TimeSlots = data.Slots

	meeting, err := b.api.CreateMeeting(ctx(), token, &data.Draft)
	if err != nil {
		log.Printf("CreateMeeting error: %v", err)
		b.conv.Clear(c.Chat().ID)
		return c.Send(fmt.Sprintf("Failed to create meeting: %s", err.Error()))
	}

	b.conv.Clear(c.Chat().ID)

	text := fmt.Sprintf(
		"\u2705 <b>Meeting created!</b>\n\n"+
			"<b>%s</b>\n"+
			"%s | %d time slot%s | %d participant%s\n\n"+
			"Share the meeting link or use the button below to view details.",
		escapeHTML(meeting.Title),
		visibilityStr(meeting.IsPublic),
		len(data.Draft.TimeSlots), pluralS(len(data.Draft.TimeSlots)),
		len(data.Draft.ParticipantIDs), pluralS(len(data.Draft.ParticipantIDs)),
	)

	markup := &tele.ReplyMarkup{}
	markup.Inline(
		markup.Row(markup.Data("\U0001f4cb View Meeting", "view", meeting.ID)),
		markup.Row(markup.Data("\U0001f4cb My Meetings", "menu", "meetings")),
	)

	return c.Send(text, &tele.SendOptions{ParseMode: tele.ModeHTML}, markup)
}

func parseTimeSlot(input string) (*apiclient.TimeSlotInput, error) {
	// Formats supported:
	// "2026-03-01 10:00 - 11:00"  (same day, short end time)
	// "2026-03-01 10:00 - 2026-03-01 11:00" (full range)
	input = strings.TrimSpace(input)

	parts := strings.SplitN(input, " - ", 2)
	if len(parts) != 2 {
		parts = strings.SplitN(input, "-", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("use format: YYYY-MM-DD HH:MM - HH:MM")
		}
	}

	startStr := strings.TrimSpace(parts[0])
	endStr := strings.TrimSpace(parts[1])

	startTime, err := time.Parse("2006-01-02 15:04", startStr)
	if err != nil {
		return nil, fmt.Errorf("invalid start time: %s", startStr)
	}

	// Try parsing end as full datetime first
	endTime, err := time.Parse("2006-01-02 15:04", endStr)
	if err != nil {
		// Try as time only (same day)
		endTime, err = time.Parse("15:04", endStr)
		if err != nil {
			return nil, fmt.Errorf("invalid end time: %s", endStr)
		}
		endTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(),
			endTime.Hour(), endTime.Minute(), 0, 0, time.UTC)
	}

	if !endTime.After(startTime) {
		return nil, fmt.Errorf("end time must be after start time")
	}

	return &apiclient.TimeSlotInput{
		StartTime: startTime.Format(time.RFC3339),
		EndTime:   endTime.Format(time.RFC3339),
	}, nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "\u2026"
}
