// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package badges

import (
	"sync"

	"encoding/json"

	"github.com/keybase/client/go/gregor"
	"github.com/keybase/client/go/logger"
	"github.com/keybase/client/go/protocol/chat1"
	"github.com/keybase/client/go/protocol/keybase1"
	jsonw "github.com/keybase/go-jsonw"
)

// BadgeState represents the number of badges on the app. It's threadsafe.
// Useable from both the client service and gregor server.
// See service:Badger for the service part that owns this.
type BadgeState struct {
	sync.Mutex

	log   logger.Logger
	state keybase1.BadgeState

	inboxVers chat1.InboxVers
	// Map from ConversationID.String to BadgeConversationInfo.
	chatUnreadMap map[string]keybase1.BadgeConversationInfo
}

// NewBadgeState creates a new empty BadgeState.
func NewBadgeState(log logger.Logger) *BadgeState {
	return &BadgeState{
		log:           log,
		inboxVers:     chat1.InboxVers(0),
		chatUnreadMap: make(map[string]keybase1.BadgeConversationInfo),
	}
}

// Exports the state summary
func (b *BadgeState) Export() (keybase1.BadgeState, error) {
	b.Lock()
	defer b.Unlock()

	b.state.Conversations = []keybase1.BadgeConversationInfo{}
	for _, info := range b.chatUnreadMap {
		b.state.Conversations = append(b.state.Conversations, info)
	}
	b.state.InboxVers = int(b.inboxVers)

	return b.state, nil
}

type problemSetBody struct {
	Count int `json:"count"`
}

type newTeamBody struct {
	TeamID string `json:"id"`
}

// UpdateWithGregor updates the badge state from a gregor state.
func (b *BadgeState) UpdateWithGregor(gstate gregor.State) error {
	b.Lock()
	defer b.Unlock()

	b.state.NewTlfs = 0
	b.state.NewFollowers = 0
	b.state.RekeysNeeded = 0
	b.state.NewGitRepoGlobalUniqueIDs = []string{}
	b.state.NewTeamIDs = nil

	items, err := gstate.Items()
	if err != nil {
		return err
	}
	for _, item := range items {
		categoryObj := item.Category()
		if categoryObj == nil {
			continue
		}
		category := categoryObj.String()
		switch category {
		case "tlf":
			jsw, err := jsonw.Unmarshal(item.Body().Bytes())
			if err != nil {
				b.log.Warning("BadgeState encountered non-json 'tlf' item: %v", err)
				continue
			}
			itemType, err := jsw.AtKey("type").GetString()
			if err != nil {
				b.log.Warning("BadgeState encountered gregor 'tlf' item without 'type': %v", err)
				continue
			}
			if itemType != "created" {
				continue
			}
			b.state.NewTlfs++
		case "kbfs_tlf_problem_set_count", "kbfs_tlf_sbs_problem_set_count":
			var body problemSetBody
			if err := json.Unmarshal(item.Body().Bytes(), &body); err != nil {
				b.log.Warning("BadgeState encountered non-json 'problem set' item: %v", err)
				continue
			}
			b.state.RekeysNeeded += body.Count
		case "follow":
			b.state.NewFollowers++
		case "new_git_repo":
			jsw, err := jsonw.Unmarshal(item.Body().Bytes())
			if err != nil {
				b.log.Warning("BadgeState encountered non-json 'new_git_repo' item: %v", err)
				continue
			}
			globalUniqueID, err := jsw.AtKey("global_unique_id").GetString()
			if err != nil {
				b.log.Warning("BadgeState encountered gregor 'new_git_repo' item without 'global_unique_id': %v", err)
				continue
			}
			b.state.NewGitRepoGlobalUniqueIDs = append(b.state.NewGitRepoGlobalUniqueIDs, globalUniqueID)
		case "team.newly_added_to_team":
			var body []newTeamBody
			if err := json.Unmarshal(item.Body().Bytes(), &body); err != nil {
				b.log.Warning("BadgeState unmarshal error for team.newly_added_to_team item: %v", err)
				continue
			}
			for _, x := range body {
				teamID, err := keybase1.TeamIDFromString(x.TeamID)
				if err != nil {
					b.log.Warning("BadgeState invalid team id in team.newly_added_to_team item: %v", err)
					continue
				}
				b.state.NewTeamIDs = append(b.state.NewTeamIDs, teamID)
			}
		case "team.request_access":
			var body []newTeamBody
			if err := json.Unmarshal(item.Body().Bytes(), &body); err != nil {
				b.log.Warning("BadgeState unmarshal error for team.request_access item: %v", err)
				continue
			}
			for _, x := range body {
				teamID, err := keybase1.TeamIDFromString(x.TeamID)
				if err != nil {
					b.log.Warning("BadgeState invalid team id in team.request_access item: %v", err)
					continue
				}
				b.state.NewTeamAccessRequests = append(b.state.NewTeamAccessRequests, teamID)
			}
		}
	}

	return nil
}

func (b *BadgeState) UpdateWithChat(update chat1.UnreadUpdate, inboxVers chat1.InboxVers) {
	b.Lock()
	defer b.Unlock()

	// Skip stale updates
	if inboxVers < b.inboxVers {
		return
	}

	b.inboxVers = inboxVers
	b.updateWithChat(update)
}

func (b *BadgeState) UpdateWithChatFull(update chat1.UnreadUpdateFull) {
	b.Lock()
	defer b.Unlock()

	if update.Ignore {
		return
	}

	// Skip stale updates
	if update.InboxVers < b.inboxVers {
		return
	}

	switch update.InboxSyncStatus {
	case chat1.SyncInboxResType_CURRENT:
	case chat1.SyncInboxResType_INCREMENTAL:
	case chat1.SyncInboxResType_CLEAR:
		b.chatUnreadMap = make(map[string]keybase1.BadgeConversationInfo)
	}

	for _, upd := range update.Updates {
		b.updateWithChat(upd)
	}

	b.inboxVers = update.InboxVers
}

func (b *BadgeState) Clear() {
	b.Lock()
	defer b.Unlock()

	b.state = keybase1.BadgeState{}
	b.inboxVers = chat1.InboxVers(0)
	b.chatUnreadMap = make(map[string]keybase1.BadgeConversationInfo)
}

func (b *BadgeState) updateWithChat(update chat1.UnreadUpdate) {
	b.chatUnreadMap[update.ConvID.String()] = keybase1.BadgeConversationInfo{
		ConvID:         keybase1.ChatConversationID(update.ConvID),
		UnreadMessages: update.UnreadMessages,
		BadgeCounts:    update.UnreadNotifyingMessages,
	}
}
