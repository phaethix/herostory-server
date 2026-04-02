package game

import (
	"herostory-server/internal/pb"
	"time"
)

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
