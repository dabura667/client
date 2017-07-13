// Auto-generated by avdl-compiler v1.3.17 (https://github.com/keybase/node-avdl-compiler)
//   Input file: avdl/keybase1/notify_badges.avdl

package keybase1

import (
	"github.com/keybase/go-framed-msgpack-rpc/rpc"
	context "golang.org/x/net/context"
)

type ChatConversationID []byte

func (o ChatConversationID) DeepCopy() ChatConversationID {
	return (func(x []byte) []byte {
		if x == nil {
			return nil
		}
		return append([]byte(nil), x...)
	})(o)
}

type BadgeState struct {
	NewTlfs       int                     `codec:"newTlfs" json:"newTlfs"`
	RekeysNeeded  int                     `codec:"rekeysNeeded" json:"rekeysNeeded"`
	NewFollowers  int                     `codec:"newFollowers" json:"newFollowers"`
	InboxVers     int                     `codec:"inboxVers" json:"inboxVers"`
	Conversations []BadgeConversationInfo `codec:"conversations" json:"conversations"`
}

func (o BadgeState) DeepCopy() BadgeState {
	return BadgeState{
		NewTlfs:      o.NewTlfs,
		RekeysNeeded: o.RekeysNeeded,
		NewFollowers: o.NewFollowers,
		InboxVers:    o.InboxVers,
		Conversations: (func(x []BadgeConversationInfo) []BadgeConversationInfo {
			var ret []BadgeConversationInfo
			for _, v := range x {
				vCopy := v.DeepCopy()
				ret = append(ret, vCopy)
			}
			return ret
		})(o.Conversations),
	}
}

type BadgeConversationInfo struct {
	ConvID            ChatConversationID `codec:"convID" json:"convID"`
	BadgeCounts       map[DeviceType]int `codec:"badgeCounts" json:"badgeCounts"`
	HasUnreadMessages bool               `codec:"hasUnreadMessages" json:"hasUnreadMessages"`
}

func (o BadgeConversationInfo) DeepCopy() BadgeConversationInfo {
	return BadgeConversationInfo{
		ConvID: o.ConvID.DeepCopy(),
		BadgeCounts: (func(x map[DeviceType]int) map[DeviceType]int {
			ret := make(map[DeviceType]int)
			for k, v := range x {
				kCopy := k.DeepCopy()
				vCopy := v
				ret[kCopy] = vCopy
			}
			return ret
		})(o.BadgeCounts),
		HasUnreadMessages: o.HasUnreadMessages,
	}
}

type BadgeStateArg struct {
	BadgeState BadgeState `codec:"badgeState" json:"badgeState"`
}

func (o BadgeStateArg) DeepCopy() BadgeStateArg {
	return BadgeStateArg{
		BadgeState: o.BadgeState.DeepCopy(),
	}
}

type NotifyBadgesInterface interface {
	BadgeState(context.Context, BadgeState) error
}

func NotifyBadgesProtocol(i NotifyBadgesInterface) rpc.Protocol {
	return rpc.Protocol{
		Name: "keybase.1.NotifyBadges",
		Methods: map[string]rpc.ServeHandlerDescription{
			"badgeState": {
				MakeArg: func() interface{} {
					ret := make([]BadgeStateArg, 1)
					return &ret
				},
				Handler: func(ctx context.Context, args interface{}) (ret interface{}, err error) {
					typedArgs, ok := args.(*[]BadgeStateArg)
					if !ok {
						err = rpc.NewTypeError((*[]BadgeStateArg)(nil), args)
						return
					}
					err = i.BadgeState(ctx, (*typedArgs)[0].BadgeState)
					return
				},
				MethodType: rpc.MethodNotify,
			},
		},
	}
}

type NotifyBadgesClient struct {
	Cli rpc.GenericClient
}

func (c NotifyBadgesClient) BadgeState(ctx context.Context, badgeState BadgeState) (err error) {
	__arg := BadgeStateArg{BadgeState: badgeState}
	err = c.Cli.Notify(ctx, "keybase.1.NotifyBadges.badgeState", []interface{}{__arg})
	return
}
