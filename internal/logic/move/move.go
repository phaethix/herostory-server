package move

import (
	"herostory-server/internal/game"
	"herostory-server/internal/pb"

	"github.com/rs/zerolog/log"
)

// Apply executes a user move based on the given command.
// Returns the UserMoveToResult ready for broadcast, or nil if the user is not online.
func Apply(uid int, cmd *pb.UserMoveToCmd) *pb.UserMoveToResult {
	user := game.GetOnlineUser(uid)
	if user == nil {
		log.Warn().
			Int("userId", uid).
			Msg("user move ignored: not online")
		return nil
	}

	user.MoveState = game.NewMoveState(
		cmd.MoveFromPosX,
		cmd.MoveFromPosY,
		cmd.MoveToPosX,
		cmd.MoveToPosY,
	)

	return &pb.UserMoveToResult{
		MoveUserId:    uint32(uid),
		MoveFromPosX:  user.MoveState.FromPosX,
		MoveFromPosY:  user.MoveState.FromPosY,
		MoveToPosX:    user.MoveState.ToPosX,
		MoveToPosY:    user.MoveState.ToPosY,
		MoveStartTime: user.MoveState.StartTime,
	}
}

// Stop parks the user at their interpolated position and returns the
// result to broadcast. It returns nil when the user is not online.
func Stop(uid int) *pb.UserStopResult {
	user := game.GetOnlineUser(uid)
	if user == nil {
		log.Warn().
			Int("userId", uid).
			Msg("user stop ignored: not online")
		return nil
	}

	x, y := user.MoveState.CurrentPos()
	user.MoveState = game.NewMoveState(x, y, x, y)

	return &pb.UserStopResult{
		StopUserId: uint32(uid),
		StopAtPosX: x,
		StopAtPosY: y,
	}
}
