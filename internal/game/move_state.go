package game

import (
	"math"
	"time"

	"herostory-server/internal/pb"
)

// MoveSpeed is pixels per second, matching the demo client's walk speed.
const MoveSpeed float32 = 300

// MoveState holds the movement state of a user.
type MoveState struct {
	FromPosX  float32
	FromPosY  float32
	ToPosX    float32
	ToPosY    float32
	StartTime uint64
}

// NewMoveState creates a MoveState from the given positions, stamped with the current time.
func NewMoveState(fromX, fromY, toX, toY float32) *MoveState {
	return &MoveState{
		FromPosX:  fromX,
		FromPosY:  fromY,
		ToPosX:    toX,
		ToPosY:    toY,
		StartTime: uint64(time.Now().UnixMilli()),
	}
}

// CurrentPos returns the interpolated position at now, based on MoveSpeed.
// A nil receiver or a completed (or zero-length) travel yields the destination.
func (ms *MoveState) CurrentPos() (x, y float32) {
	if ms == nil {
		return 0, 0
	}

	dx := float64(ms.ToPosX - ms.FromPosX)
	dy := float64(ms.ToPosY - ms.FromPosY)
	dist := math.Hypot(dx, dy)
	if dist < 1e-5 {
		return ms.ToPosX, ms.ToPosY
	}

	elapsed := float64(time.Now().UnixMilli()-int64(ms.StartTime)) / 1000
	if elapsed < 0 {
		elapsed = 0
	}
	travel := dist / float64(MoveSpeed)
	if elapsed >= travel {
		return ms.ToPosX, ms.ToPosY
	}

	p := elapsed / travel
	return ms.FromPosX + float32(dx*p), ms.FromPosY + float32(dy*p)
}

// ToPB converts MoveState to its protobuf representation.
func (ms *MoveState) ToPB() *pb.WhoElseIsHereResult_UserInfo_MoveState {
	return &pb.WhoElseIsHereResult_UserInfo_MoveState{
		FromPosX:  ms.FromPosX,
		FromPosY:  ms.FromPosY,
		ToPosX:    ms.ToPosX,
		ToPosY:    ms.ToPosY,
		StartTime: ms.StartTime,
	}
}
