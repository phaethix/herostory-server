// Package attack contains the attack-related game logic.
//
// The package is deliberately decoupled from networking: callers pass in a
// command and receive a value-typed Result describing what the network
// layer should do (broadcasts, persistence is registered internally).
package attack

import (
	"google.golang.org/protobuf/reflect/protoreflect"

	"herostory-server/internal/game"
	"herostory-server/internal/pb"
	"herostory-server/internal/userpersist"

	"github.com/rs/zerolog/log"
)

// subtractHp is the fixed HP loss per attack.
// In a real game this should be derived from attacker / target attributes.
const subtractHp int32 = 10

// Result is what a single attack produces. The Broadcasts slice is
// pre-ordered: AttkResult, then SubtractHpResult, then optional DieResult.
// Callers should iterate it in order rather than reading the typed
// fields, so message ordering is owned by this package.
type Result struct {
	Broadcasts []protoreflect.ProtoMessage
}

// Apply executes an attack from attackerID against the target referenced
// by cmd. On success it mutates the target's in-memory HP, registers a
// delayed persistence task, and returns the messages to broadcast. It
// returns nil when the attack is invalid (offline attacker / target,
// self-attack, or already-dead target) — in which case nothing is
// mutated and no message should be sent.
func Apply(attackerID int, cmd *pb.UserAttkCmd) *Result {
	target := validate(attackerID, cmd)
	if target == nil {
		return nil
	}

	targetID := target.UserID
	target.CurrHp -= subtractHp

	// Schedule a delayed write of the new HP. Repeated attacks on the same
	// target within the lazy-save quiet window collapse into one SQL
	// update.
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
	}

	return r
}

// validate enforces every precondition of an attack. It returns the
// target OnlineUser on success, or nil if the attack must be ignored.
// Splitting validation out keeps Apply readable and makes the rules
// trivially unit-testable.
func validate(attackerID int, cmd *pb.UserAttkCmd) *game.OnlineUser {
	if cmd == nil || attackerID <= 0 {
		return nil
	}

	targetID := int(cmd.GetTargetUserId())
	if targetID <= 0 {
		return nil
	}
	if targetID == attackerID {
		log.Debug().
			Int("userId", attackerID).
			Msg("attack ignored: self-attack")
		return nil
	}

	if game.GetOnlineUser(attackerID) == nil {
		log.Warn().
			Int("userId", attackerID).
			Msg("attack ignored: attacker not online")
		return nil
	}

	target := game.GetOnlineUser(targetID)
	if target == nil {
		log.Warn().
			Int("targetUserId", targetID).
			Msg("attack ignored: target not online")
		return nil
	}
	if target.CurrHp <= 0 {
		// Already dead, nothing to do. Not logged: this can be a normal
		// race when several attackers hit the same dying target.
		return nil
	}

	return target
}
