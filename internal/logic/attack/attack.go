// Package attack is deliberately unaware of sockets: Apply mutates
// in-memory HP and returns the messages to broadcast, in order.
package attack

import (
	"herostory-server/internal/game"
	"herostory-server/internal/pb"
	"herostory-server/internal/userpersist"

	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// subtractHp is fixed because this demo has no attacker/target stats.
const subtractHp int32 = 10

// Result.Broadcasts is pre-ordered: Attk, SubtractHp, then optional Die.
// Iterate the slice; do not reconstruct the sequence at the call site.
// LoserID is non-zero when this hit killed the target.
type Result struct {
	Broadcasts []protoreflect.ProtoMessage
	WinnerID   int
	LoserID    int
}

// Apply returns nil when the attack is invalid; nothing is mutated in that case.
func Apply(attackerID int, cmd *pb.UserAttkCmd) *Result {
	target := validate(attackerID, cmd)
	if target == nil {
		return nil
	}

	targetID := target.UserID
	target.CurrHp -= subtractHp
	userpersist.SaveOrUpdate(target)

	r := &Result{
		Broadcasts: []protoreflect.ProtoMessage{
			&pb.UserAttkResult{
				AttkUserId:   uint32(attackerID),
				TargetUserId: uint32(targetID),
			},
			&pb.UserSubtractHpResult{
				TargetUserId: uint32(targetID),
				SubtractHp:   uint32(subtractHp),
			},
		},
	}
	if target.CurrHp <= 0 {
		r.Broadcasts = append(r.Broadcasts, &pb.UserDieResult{
			TargetUserId: uint32(targetID),
		})
		r.WinnerID = attackerID
		r.LoserID = targetID
	}
	return r
}

func validate(attackerID int, cmd *pb.UserAttkCmd) *game.OnlineUser {
	if cmd == nil || attackerID <= 0 {
		return nil
	}

	targetID := int(cmd.TargetUserId)
	if targetID <= 0 {
		return nil
	}
	if targetID == attackerID {
		log.Debug().Int("userId", attackerID).Msg("attack ignored: self-attack")
		return nil
	}

	if game.GetOnlineUser(attackerID) == nil {
		log.Warn().Int("userId", attackerID).Msg("attack ignored: attacker not online")
		return nil
	}

	target := game.GetOnlineUser(targetID)
	if target == nil {
		log.Warn().Int("targetUserId", targetID).Msg("attack ignored: target not online")
		return nil
	}
	if target.CurrHp <= 0 {
		// Several attackers can hit the same dying target in one tick.
		return nil
	}
	return target
}
