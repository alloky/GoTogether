package bot

import (
	"fmt"
	"log"
	"strconv"

	"github.com/gotogether/tgbot/internal/apiclient"
	tele "gopkg.in/telebot.v3"
)

// cbView handles "view:{meetingID}" — show meeting detail
func (b *Bot) cbView(c tele.Context) error {
	meetingID := c.Callback().Data
	_ = c.Respond()
	return b.sendMeetingDetail(c, meetingID)
}

func (b *Bot) sendMeetingDetail(c tele.Context, meetingID string) error {
	token, err := b.ensureAuth(c)
	if err != nil {
		return c.Send("Please /start first.")
	}

	meeting, err := b.api.GetMeeting(ctx(), token, meetingID)
	if err != nil {
		log.Printf("GetMeeting error: %v", err)
		return c.Send("Failed to load meeting. It may have been deleted.")
	}

	// Determine user's role from the auth response
	userID := b.resolveUserID(c.Sender().ID, meeting)

	text := renderMeetingDetail(meeting, userID)
	markup := b.buildDetailKeyboard(meeting, userID)

	return c.Send(text, &tele.SendOptions{ParseMode: tele.ModeHTML}, markup)
}

// resolveUserID tries to find the backend user ID for the given telegram user.
func (b *Bot) resolveUserID(telegramID int64, meeting *apiclient.Meeting) string {
	email := fmt.Sprintf("tg_%d@telegram.local", telegramID)
	if meeting.Organizer != nil && meeting.Organizer.Email == email {
		return meeting.OrganizerID
	}
	for _, p := range meeting.Participants {
		if p.User != nil && p.User.Email == email {
			return p.UserID
		}
	}
	return "unknown"
}

func (b *Bot) buildDetailKeyboard(m *apiclient.Meeting, userID string) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}
	var rows []tele.Row

	isOrganizer := m.OrganizerID == userID
	isParticipant := false
	for _, p := range m.Participants {
		if p.UserID == userID {
			isParticipant = true
			break
		}
	}
	isPending := m.Status == "pending"

	if isPending && len(m.TimeSlots) > 0 {
		rows = append(rows, markup.Row(
			markup.Data("\U0001f5f3 Vote on Times", "vote", m.ID),
		))
	}

	if isParticipant && isPending {
		rows = append(rows, markup.Row(
			markup.Data("\u2705 Accept", "rsvp", m.ID+"|accepted"),
			markup.Data("\u274c Decline", "rsvp", m.ID+"|declined"),
		))
	}

	if isOrganizer && isPending && len(m.TimeSlots) > 0 {
		rows = append(rows, markup.Row(
			markup.Data("\U0001f3af Confirm Meeting", "conf", m.ID),
		))
	}

	if isOrganizer {
		rows = append(rows, markup.Row(
			markup.Data("\U0001f5d1 Delete Meeting", "del", m.ID),
		))
	}

	rows = append(rows, markup.Row(
		markup.Data("\u2b05 Back to Meetings", "back", "meetings"),
	))

	markup.Inline(rows...)
	return markup
}

// cbVote — show vote interface
func (b *Bot) cbVote(c tele.Context) error {
	meetingID := c.Callback().Data
	_ = c.Respond()

	token, err := b.ensureAuth(c)
	if err != nil {
		return c.Send("Please /start first.")
	}

	meeting, err := b.api.GetMeeting(ctx(), token, meetingID)
	if err != nil {
		return c.Send("Failed to load meeting.")
	}

	// Pre-select slots the user already voted for
	email := fmt.Sprintf("tg_%d@telegram.local", c.Sender().ID)
	preselected := make(map[string]bool)
	for _, slot := range meeting.TimeSlots {
		for _, voter := range slot.Voters {
			if voter.Email == email {
				preselected[slot.ID] = true
				break
			}
		}
	}

	b.votes.Init(c.Chat().ID, meetingID, preselected)
	return b.sendVoteView(c, meeting)
}

func (b *Bot) sendVoteView(c tele.Context, meeting *apiclient.Meeting) error {
	selected := b.votes.Get(c.Chat().ID, meeting.ID)
	text := renderVoteView(meeting, selected)

	markup := &tele.ReplyMarkup{}
	var rows []tele.Row
	compactMID := compactUUID(meeting.ID)

	for i, slot := range meeting.TimeSlots {
		icon := "\u2b1c"
		if selected[slot.ID] {
			icon = "\u2705"
		}
		rows = append(rows, markup.Row(
			markup.Data(
				fmt.Sprintf("%s %s\u2013%s", icon, fmtTime(slot.StartTime), fmtTimeShort(slot.EndTime)),
				"vtog",
				compactMID+"|"+strconv.Itoa(i),
			),
		))
	}

	rows = append(rows,
		markup.Row(
			markup.Data("\U0001f4e8 Submit Votes", "vsub", meeting.ID),
			markup.Data("\u274c Cancel", "view", meeting.ID),
		),
	)

	markup.Inline(rows...)

	return c.Send(text, &tele.SendOptions{ParseMode: tele.ModeHTML}, markup)
}

// cbVoteToggle — toggle a single slot selection
func (b *Bot) cbVoteToggle(c tele.Context) error {
	data := c.Callback().Data
	_ = c.Respond()

	parts := splitPipe(data)
	if len(parts) != 2 {
		return nil
	}
	compactMID := parts[0]
	slotIdx, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil
	}
	meetingID := expandUUID(compactMID)

	// Re-fetch meeting to resolve slot index → slot ID
	token, authErr := b.ensureAuth(c)
	if authErr != nil {
		return nil
	}
	meeting, fetchErr := b.api.GetMeeting(ctx(), token, meetingID)
	if fetchErr != nil {
		return nil
	}
	if slotIdx < 0 || slotIdx >= len(meeting.TimeSlots) {
		return nil
	}
	slotID := meeting.TimeSlots[slotIdx].ID

	b.votes.Toggle(c.Chat().ID, meetingID, slotID)

	selected := b.votes.Get(c.Chat().ID, meetingID)
	text := renderVoteView(meeting, selected)

	markup := &tele.ReplyMarkup{}
	var rows []tele.Row
	for i, slot := range meeting.TimeSlots {
		icon := "\u2b1c"
		if selected[slot.ID] {
			icon = "\u2705"
		}
		rows = append(rows, markup.Row(
			markup.Data(
				fmt.Sprintf("%s %s\u2013%s", icon, fmtTime(slot.StartTime), fmtTimeShort(slot.EndTime)),
				"vtog",
				compactMID+"|"+strconv.Itoa(i),
			),
		))
	}
	rows = append(rows,
		markup.Row(
			markup.Data("\U0001f4e8 Submit Votes", "vsub", meetingID),
			markup.Data("\u274c Cancel", "view", meetingID),
		),
	)
	markup.Inline(rows...)

	return c.Edit(text, &tele.SendOptions{ParseMode: tele.ModeHTML}, markup)
}

// cbVoteSubmit — submit the votes
func (b *Bot) cbVoteSubmit(c tele.Context) error {
	meetingID := c.Callback().Data
	_ = c.Respond()

	token, err := b.ensureAuth(c)
	if err != nil {
		return c.Send("Please /start first.")
	}

	slotIDs := b.votes.GetSelected(c.Chat().ID, meetingID)
	if len(slotIDs) == 0 {
		return c.Send("No slots selected. Tap slots to toggle, then submit.")
	}

	err = b.api.Vote(ctx(), token, meetingID, slotIDs)
	if err != nil {
		log.Printf("Vote error: %v", err)
		return c.Send(fmt.Sprintf("Failed to submit votes: %s", err.Error()))
	}

	b.votes.Clear(c.Chat().ID, meetingID)

	_ = c.Send(fmt.Sprintf("\u2705 Votes submitted! (%d slot%s)", len(slotIDs), pluralS(len(slotIDs))))
	return b.sendMeetingDetail(c, meetingID)
}

// cbRSVP — accept or decline
func (b *Bot) cbRSVP(c tele.Context) error {
	data := c.Callback().Data
	_ = c.Respond()

	parts := splitPipe(data)
	if len(parts) != 2 {
		return nil
	}
	meetingID, status := parts[0], parts[1]

	token, err := b.ensureAuth(c)
	if err != nil {
		return c.Send("Please /start first.")
	}

	err = b.api.UpdateRSVP(ctx(), token, meetingID, status)
	if err != nil {
		log.Printf("RSVP error: %v", err)
		return c.Send(fmt.Sprintf("Failed to update RSVP: %s", err.Error()))
	}

	icon := "\u2705"
	if status == "declined" {
		icon = "\u274c"
	}
	_ = c.Send(fmt.Sprintf("%s RSVP: %s", icon, status))
	return b.sendMeetingDetail(c, meetingID)
}

// cbConfirm — show confirm slot picker
func (b *Bot) cbConfirm(c tele.Context) error {
	meetingID := c.Callback().Data
	_ = c.Respond()

	token, err := b.ensureAuth(c)
	if err != nil {
		return c.Send("Please /start first.")
	}

	meeting, err := b.api.GetMeeting(ctx(), token, meetingID)
	if err != nil {
		return c.Send("Failed to load meeting.")
	}

	text := renderConfirmView(meeting)
	markup := &tele.ReplyMarkup{}
	var rows []tele.Row

	compactMID := compactUUID(meetingID)

	rows = append(rows, markup.Row(
		markup.Data("\U0001f3b2 Auto-pick Best", "cfsl", compactMID+"|auto"),
	))

	for i, slot := range meeting.TimeSlots {
		rows = append(rows, markup.Row(
			markup.Data(
				fmt.Sprintf("%s\u2013%s (%d vote%s)",
					fmtTime(slot.StartTime),
					fmtTimeShort(slot.EndTime),
					slot.VoteCount,
					pluralS(slot.VoteCount),
				),
				"cfsl",
				compactMID+"|"+strconv.Itoa(i),
			),
		))
	}

	rows = append(rows, markup.Row(
		markup.Data("\u274c Cancel", "view", meetingID),
	))

	markup.Inline(rows...)
	return c.Send(text, &tele.SendOptions{ParseMode: tele.ModeHTML}, markup)
}

// cbConfirmSlot — confirm with selected slot
func (b *Bot) cbConfirmSlot(c tele.Context) error {
	data := c.Callback().Data
	_ = c.Respond()

	parts := splitPipe(data)
	if len(parts) != 2 {
		return nil
	}
	meetingID := expandUUID(parts[0])
	slotChoice := parts[1]

	token, err := b.ensureAuth(c)
	if err != nil {
		return c.Send("Please /start first.")
	}

	var slotID *string
	if slotChoice != "auto" {
		// Resolve slot index to slot ID by re-fetching the meeting
		slotIdx, err := strconv.Atoi(slotChoice)
		if err != nil {
			return c.Send("Invalid slot selection.")
		}
		meeting, err := b.api.GetMeeting(ctx(), token, meetingID)
		if err != nil {
			return c.Send("Failed to load meeting.")
		}
		if slotIdx < 0 || slotIdx >= len(meeting.TimeSlots) {
			return c.Send("Invalid slot selection.")
		}
		slotID = &meeting.TimeSlots[slotIdx].ID
	}

	_, err = b.api.ConfirmMeeting(ctx(), token, meetingID, slotID)
	if err != nil {
		log.Printf("Confirm error: %v", err)
		return c.Send(fmt.Sprintf("Failed to confirm: %s", err.Error()))
	}

	_ = c.Send("\u2705 Meeting confirmed!")
	return b.sendMeetingDetail(c, meetingID)
}

// cbDelete — delete a meeting
func (b *Bot) cbDelete(c tele.Context) error {
	meetingID := c.Callback().Data
	_ = c.Respond()

	token, err := b.ensureAuth(c)
	if err != nil {
		return c.Send("Please /start first.")
	}

	err = b.api.DeleteMeeting(ctx(), token, meetingID)
	if err != nil {
		log.Printf("Delete error: %v", err)
		return c.Send(fmt.Sprintf("Failed to delete: %s", err.Error()))
	}

	_ = c.Send("\U0001f5d1 Meeting deleted.")
	return b.sendMeetingsList(c, 0)
}

// cbMeetingsPage — pagination for meetings list
func (b *Bot) cbMeetingsPage(c tele.Context) error {
	page, _ := strconv.Atoi(c.Callback().Data)
	_ = c.Respond()
	return b.sendMeetingsList(c, page)
}

// cbBack — navigate back
func (b *Bot) cbBack(c tele.Context) error {
	target := c.Callback().Data
	_ = c.Respond()

	switch target {
	case "meetings":
		return b.sendMeetingsList(c, 0)
	case "calendar":
		return b.sendCalendar(c, "")
	case "main":
		return b.handleStart(c)
	default:
		return b.sendMeetingsList(c, 0)
	}
}

// cbCalendarTag — filter calendar by tag
func (b *Bot) cbCalendarTag(c tele.Context) error {
	tag := c.Callback().Data
	_ = c.Respond()

	if tag == "_clear" {
		tag = ""
	}
	return b.sendCalendar(c, tag)
}

// cbMenu — main menu buttons
func (b *Bot) cbMenu(c tele.Context) error {
	action := c.Callback().Data
	_ = c.Respond()

	switch action {
	case "meetings":
		return b.sendMeetingsList(c, 0)
	case "calendar":
		return b.sendCalendar(c, "")
	case "new":
		return b.handleNew(c)
	case "help":
		return b.handleHelp(c)
	default:
		return nil
	}
}

// cbVisibility — handle visibility selection during meeting creation
func (b *Bot) cbVisibility(c tele.Context) error {
	d := b.conv.Get(c.Chat().ID)
	if d.State != StateAwaitVisibility {
		_ = c.Respond()
		return nil
	}
	d.Draft.IsPublic = c.Callback().Data == "public"
	d.State = StateAwaitTimeSlots
	b.conv.Set(c.Chat().ID, d)
	_ = c.Respond()
	return c.Send(
		fmt.Sprintf("Visibility: %s\n\n"+
			"Step 4/6: Enter <b>time slots</b>, one per message.\n"+
			"Format: <code>YYYY-MM-DD HH:MM - HH:MM</code>\n"+
			"Example: <code>2026-03-01 10:00 - 11:00</code>\n\n"+
			"Send /done when finished.",
			visibilityStr(d.Draft.IsPublic)),
		&tele.SendOptions{ParseMode: tele.ModeHTML},
	)
}

// cbAddParticipant — handle participant selection during meeting creation
func (b *Bot) cbAddParticipant(c tele.Context) error {
	d := b.conv.Get(c.Chat().ID)
	uid := c.Callback().Data
	_ = c.Respond()

	if uid == "_done" {
		return b.createMeetingFromDraft(c, d)
	}

	// Add participant (avoid duplicates)
	for _, existing := range d.Draft.ParticipantIDs {
		if existing == uid {
			return c.Send("Already added. Search another name or /done.")
		}
	}
	d.Draft.ParticipantIDs = append(d.Draft.ParticipantIDs, uid)
	d.State = StateAwaitMoreParticipants
	b.conv.Set(c.Chat().ID, d)
	return c.Send(fmt.Sprintf(
		"\u2705 Participant added (%d total). Search another or /done.",
		len(d.Draft.ParticipantIDs),
	))
}

// splitPipe splits "a|b" into ["a", "b"]
func splitPipe(s string) []string {
	result := []string{}
	current := ""
	for _, ch := range s {
		if ch == '|' {
			result = append(result, current)
			current = ""
		} else {
			current += string(ch)
		}
	}
	result = append(result, current)
	return result
}

// compactUUID removes hyphens from a UUID to save space in callback_data.
// "62a442f7-32fb-4171-bc4e-2ae27813d526" → "62a442f732fb4171bc4e2ae27813d526"
func compactUUID(uuid string) string {
	result := make([]byte, 0, 32)
	for i := 0; i < len(uuid); i++ {
		if uuid[i] != '-' {
			result = append(result, uuid[i])
		}
	}
	return string(result)
}

// expandUUID re-inserts hyphens into a compacted UUID.
// "62a442f732fb4171bc4e2ae27813d526" → "62a442f7-32fb-4171-bc4e-2ae27813d526"
func expandUUID(s string) string {
	if len(s) != 32 {
		return s // not a compacted UUID, return as-is
	}
	return s[0:8] + "-" + s[8:12] + "-" + s[12:16] + "-" + s[16:20] + "-" + s[20:32]
}
