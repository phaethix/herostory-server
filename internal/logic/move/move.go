package move

import (
	"herostory-server/internal/game"
	"herostory-server/internal/pb"

	"github.com/rs/zerolog/log"
)

// Apply records a move and returns the result to broadcast, or nil if the user is offline.
func Apply(uid int, cmd *pb.UserMoveToCmd) *pb.UserMoveToResult {
	user := game.GetOnlineUser(uid)
	if user == nil {
		log.Warn().Int("userId", uid).Msg("user move ignored: not online")
		return nil
	}

	user.MoveState = game.NewMoveState(
		cmd.MoveFromPosX,
		cmd.MoveFromPosY,
		cmd.MoveToPosX,
		cmd.MoveToPosY,
	)
	ms := user.MoveState
	return &pb.UserMoveToResult{
		MoveUserId:    uint32(uid),
		MoveFromPosX:  ms.FromPosX,
		MoveFromPosY:  ms.FromPosY,
		MoveToPosX:    ms.ToPosX,
		MoveToPosY:    ms.ToPosY,
		MoveStartTime: ms.StartTime,
	}
}

// Stop freezes the avatar at the interpolated point. Parking From=To
// keeps WhoElseIsHere from animating a leftover path.
func Stop(uid int) *pb.UserStopResult {
	user := game.GetOnlineUser(uid)
	if user == nil {
		log.Warn().Int("userId", uid).Msg("user stop ignored: not online")
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
