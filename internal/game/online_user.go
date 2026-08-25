package game

import (
	"iter"
	"maps"
)

// OnlineUser is session state. The map is unsynchronised: every
// function in this package must run on the main goroutine.
type OnlineUser struct {
	UserID     int
	UserName   string
	HeroAvatar string
	CurrHp     int32
	MoveState  *MoveState
}

var onlineUsers = make(map[int]*OnlineUser)

func AddOnlineUser(u *OnlineUser) {
	if u == nil || u.UserID <= 0 {
		return
	}
	onlineUsers[u.UserID] = u
}

func RemoveOnlineUser(userID int) {
	delete(onlineUsers, userID)
}

// GetOnlineUser returns nil when the user is not in the session map.
func GetOnlineUser(userID int) *OnlineUser {
	return onlineUsers[userID]
}

// OnlineUsers yields map order. Callers must already be on the main goroutine.
func OnlineUsers() iter.Seq[*OnlineUser] {
	return maps.Values(onlineUsers)
}
