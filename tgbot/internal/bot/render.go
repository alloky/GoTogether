package bot

import (
	"fmt"
	"strings"
	"time"

	"github.com/gotogether/tgbot/internal/apiclient"
)

func fmtTime(t time.Time) string {
	return t.Format("Mon, Jan 2 15:04")
}

func fmtDate(t time.Time) string {
	return t.Format("Mon, Jan 2, 2006")
}

func fmtTimeShort(t time.Time) string {
	return t.Format("15:04")
}

func statusEmoji(status string) string {
	switch status {
	case "pending":
		return "\u23f3" // hourglass
	case "confirmed":
		return "\u2705" // green check
	case "cancelled":
		return "\u274c" // red X
	default:
		return "\u2753" // question mark
	}
}

func visibilityStr(isPublic bool) string {
	if isPublic {
		return "\U0001f30d Public"
	}
	return "\U0001f512 Private"
}

func renderMeetingList(meetings []apiclient.Meeting, page, perPage int) string {
	if len(meetings) == 0 {
		return "You have no meetings yet.\n\nUse /new to create one!"
	}

	total := len(meetings)
	start := page * perPage
	end := start + perPage
	if end > total {
		end = total
	}
	if start >= total {
		return "No more meetings."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\U0001f4cb <b>Your Meetings</b> (%d total)\n\n", total))

	for i := start; i < end; i++ {
		m := meetings[i]
		sb.WriteString(fmt.Sprintf("%s <b>%s</b> [%s]\n", statusEmoji(m.Status), escapeHTML(m.Title), m.Status))
		if len(m.Tags) > 0 {
			sb.WriteString(fmt.Sprintf("   \U0001f3f7 %s\n", strings.Join(m.Tags, ", ")))
		}
		if !m.IsPublic {
			sb.WriteString("   \U0001f512 Private\n")
		}
		sb.WriteString("\n")
	}

	if total > perPage {
		sb.WriteString(fmt.Sprintf("Page %d/%d", page+1, (total+perPage-1)/perPage))
	}

	return sb.String()
}

func renderMeetingDetail(m *apiclient.Meeting, userID string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s <b>%s</b>\n", statusEmoji(m.Status), escapeHTML(m.Title)))
	sb.WriteString(fmt.Sprintf("Status: %s | %s\n", m.Status, visibilityStr(m.IsPublic)))

	if m.Organizer != nil {
		sb.WriteString(fmt.Sprintf("\U0001f464 Organizer: %s\n", escapeHTML(m.Organizer.DisplayName)))
	}

	if m.Description != "" {
		sb.WriteString(fmt.Sprintf("\n\U0001f4dd %s\n", escapeHTML(m.Description)))
	}

	if len(m.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("\n\U0001f3f7 Tags: %s\n", strings.Join(m.Tags, ", ")))
	}

	// Time slots
	if len(m.TimeSlots) > 0 {
		sb.WriteString("\n\U0001f4c5 <b>Time Slots:</b>\n")
		for i, slot := range m.TimeSlots {
			marker := fmt.Sprintf("%d.", i+1)
			if m.ConfirmedSlotID != nil && slot.ID == *m.ConfirmedSlotID {
				marker = "\u2705"
			}
			sb.WriteString(fmt.Sprintf("  %s %s \u2013 %s (%d vote%s)\n",
				marker,
				fmtTime(slot.StartTime),
				fmtTimeShort(slot.EndTime),
				slot.VoteCount,
				pluralS(slot.VoteCount),
			))
		}
	}

	// Participants
	if len(m.Participants) > 0 {
		sb.WriteString("\n\U0001f465 <b>Participants:</b>\n")
		for _, p := range m.Participants {
			name := "Unknown"
			if p.User != nil {
				name = p.User.DisplayName
			}
			rsvpIcon := "\u2753"
			switch p.RSVPStatus {
			case "accepted":
				rsvpIcon = "\u2705"
			case "declined":
				rsvpIcon = "\u274c"
			case "invited":
				rsvpIcon = "\U0001f4e9"
			}
			sb.WriteString(fmt.Sprintf("  %s %s\n", rsvpIcon, escapeHTML(name)))
		}
	}

	// Show user's role
	isOrganizer := m.OrganizerID == userID
	isParticipant := false
	myRSVP := ""
	for _, p := range m.Participants {
		if p.UserID == userID {
			isParticipant = true
			myRSVP = p.RSVPStatus
			break
		}
	}

	sb.WriteString("\n\u2014\u2014\u2014\n")
	if isOrganizer {
		sb.WriteString("\U0001f451 You are the organizer")
	} else if isParticipant {
		sb.WriteString(fmt.Sprintf("\U0001f4e9 Your RSVP: %s", myRSVP))
	}

	return sb.String()
}

func renderCalendar(meetings []apiclient.Meeting) string {
	// Filter to confirmed only
	type calEvent struct {
		meeting apiclient.Meeting
		slot    apiclient.TimeSlot
	}
	var events []calEvent
	for _, m := range meetings {
		if m.Status != "confirmed" || m.ConfirmedSlotID == nil {
			continue
		}
		for _, s := range m.TimeSlots {
			if s.ID == *m.ConfirmedSlotID {
				events = append(events, calEvent{meeting: m, slot: s})
				break
			}
		}
	}

	if len(events) == 0 {
		return "\U0001f4c5 <b>Calendar</b>\n\nNo upcoming confirmed events."
	}

	// Group by date
	type dateGroup struct {
		date   string
		events []calEvent
	}
	groups := []dateGroup{}
	groupMap := map[string]int{}

	for _, e := range events {
		dateKey := e.slot.StartTime.Format("2006-01-02")
		if idx, ok := groupMap[dateKey]; ok {
			groups[idx].events = append(groups[idx].events, e)
		} else {
			groupMap[dateKey] = len(groups)
			groups = append(groups, dateGroup{date: dateKey, events: []calEvent{e}})
		}
	}

	var sb strings.Builder
	sb.WriteString("\U0001f4c5 <b>Calendar</b>\n\n")

	for _, g := range groups {
		t, _ := time.Parse("2006-01-02", g.date)
		sb.WriteString(fmt.Sprintf("<b>%s</b>\n", fmtDate(t)))
		for _, e := range g.events {
			tags := ""
			if len(e.meeting.Tags) > 0 {
				tags = " [\U0001f3f7 " + strings.Join(e.meeting.Tags, ", ") + "]"
			}
			organizer := ""
			if e.meeting.Organizer != nil {
				organizer = " by " + e.meeting.Organizer.DisplayName
			}
			sb.WriteString(fmt.Sprintf("  \u2022 %s\u2013%s <b>%s</b>%s%s\n",
				fmtTimeShort(e.slot.StartTime),
				fmtTimeShort(e.slot.EndTime),
				escapeHTML(e.meeting.Title),
				tags,
				organizer,
			))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func renderVoteView(m *apiclient.Meeting, selectedSlots map[string]bool) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\U0001f5f3 <b>Vote: %s</b>\n\n", escapeHTML(m.Title)))
	sb.WriteString("Tap slots to toggle your vote, then submit.\n\n")

	for i, slot := range m.TimeSlots {
		check := "\u2b1c" // white square
		if selectedSlots[slot.ID] {
			check = "\u2705" // green check
		}
		sb.WriteString(fmt.Sprintf("%s %d. %s \u2013 %s (%d vote%s)\n",
			check, i+1,
			fmtTime(slot.StartTime),
			fmtTimeShort(slot.EndTime),
			slot.VoteCount,
			pluralS(slot.VoteCount),
		))
	}

	return sb.String()
}

func renderConfirmView(m *apiclient.Meeting) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\U0001f3af <b>Confirm: %s</b>\n\n", escapeHTML(m.Title)))
	sb.WriteString("Pick a time slot to confirm, or auto-pick the most voted one.\n\n")

	for i, slot := range m.TimeSlots {
		sb.WriteString(fmt.Sprintf("%d. %s \u2013 %s (%d vote%s)\n",
			i+1,
			fmtTime(slot.StartTime),
			fmtTimeShort(slot.EndTime),
			slot.VoteCount,
			pluralS(slot.VoteCount),
		))
	}

	return sb.String()
}

func renderHelp() string {
	return `<b>GoTogether Bot</b> - Meeting Planner

<b>Commands:</b>
/start - Welcome & register
/meetings - View your meetings
/calendar - Upcoming confirmed events
/new - Create a new meeting
/link - Link to your web account
/help - Show this help

<b>Features:</b>
• Create meetings with multiple time slots
• Invite participants by searching usernames
• Vote on preferred time slots
• Accept or decline invitations
• Organizer can confirm the final time
• Tag meetings for easy filtering
• Public & private meeting visibility
• Link your Telegram to your web account`
}

func renderSearchResults(users []apiclient.User) string {
	if len(users) == 0 {
		return "No users found."
	}
	var sb strings.Builder
	sb.WriteString("\U0001f50d <b>Search Results:</b>\n\n")
	for _, u := range users {
		sb.WriteString(fmt.Sprintf("\u2022 %s (%s)\n", escapeHTML(u.DisplayName), escapeHTML(u.Email)))
	}
	return sb.String()
}

func escapeHTML(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	return r.Replace(s)
}

func pluralS(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
