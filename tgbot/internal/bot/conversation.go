package bot

import (
	"sync"

	"github.com/gotogether/tgbot/internal/apiclient"
)

type ConvState int

const (
	StateIdle ConvState = iota
	StateAwaitTitle
	StateAwaitDescription
	StateAwaitVisibility
	StateAwaitTimeSlots
	StateAwaitMoreSlots
	StateAwaitTags
	StateAwaitParticipantSearch
	StateAwaitMoreParticipants
)

type ConvData struct {
	State  ConvState
	Draft  apiclient.CreateMeetingInput
	Slots  []apiclient.TimeSlotInput
	VoteSelections map[string]bool // meetingID:slotID -> selected
}

type ConversationManager struct {
	sessions sync.Map // int64 (chatID) -> *ConvData
}

func NewConversationManager() *ConversationManager {
	return &ConversationManager{}
}

func (cm *ConversationManager) Get(chatID int64) *ConvData {
	v, ok := cm.sessions.Load(chatID)
	if !ok {
		return &ConvData{State: StateIdle}
	}
	return v.(*ConvData)
}

func (cm *ConversationManager) Set(chatID int64, data *ConvData) {
	cm.sessions.Store(chatID, data)
}

func (cm *ConversationManager) Clear(chatID int64) {
	cm.sessions.Delete(chatID)
}

// Vote selections are stored separately, keyed by chatID
type VoteStore struct {
	mu    sync.Mutex
	store map[int64]map[string]map[string]bool // chatID -> meetingID -> slotID -> selected
}

func NewVoteStore() *VoteStore {
	return &VoteStore{store: make(map[int64]map[string]map[string]bool)}
}

func (vs *VoteStore) Toggle(chatID int64, meetingID, slotID string) bool {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	if vs.store[chatID] == nil {
		vs.store[chatID] = make(map[string]map[string]bool)
	}
	if vs.store[chatID][meetingID] == nil {
		vs.store[chatID][meetingID] = make(map[string]bool)
	}

	current := vs.store[chatID][meetingID][slotID]
	vs.store[chatID][meetingID][slotID] = !current
	return !current
}

func (vs *VoteStore) Get(chatID int64, meetingID string) map[string]bool {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	if vs.store[chatID] == nil || vs.store[chatID][meetingID] == nil {
		return make(map[string]bool)
	}

	result := make(map[string]bool)
	for k, v := range vs.store[chatID][meetingID] {
		result[k] = v
	}
	return result
}

func (vs *VoteStore) GetSelected(chatID int64, meetingID string) []string {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	var selected []string
	if vs.store[chatID] != nil && vs.store[chatID][meetingID] != nil {
		for slotID, sel := range vs.store[chatID][meetingID] {
			if sel {
				selected = append(selected, slotID)
			}
		}
	}
	return selected
}

func (vs *VoteStore) Clear(chatID int64, meetingID string) {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	if vs.store[chatID] != nil {
		delete(vs.store[chatID], meetingID)
	}
}

func (vs *VoteStore) Init(chatID int64, meetingID string, preselected map[string]bool) {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	if vs.store[chatID] == nil {
		vs.store[chatID] = make(map[string]map[string]bool)
	}
	vs.store[chatID][meetingID] = make(map[string]bool)
	for k, v := range preselected {
		vs.store[chatID][meetingID][k] = v
	}
}
