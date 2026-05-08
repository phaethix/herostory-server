package game

import (
	"iter"
	"maps"
)

// OnlineUser holds the runtime data of a logged-in user.
// All functions operating on onlineUsers are called from the main thread, no lock needed.
type OnlineUser struct {
	UserID     int
	UserName   string
	HeroAvatar string
	CurrHp     int32
	MoveState  *MoveState
}

var onlineUsers = make(map[int]*OnlineUser)

// AddOnlineUser registers a user as online.
func AddOnlineUser(u *OnlineUser) {
	if u == nil || u.UserID <= 0 {
		return
	}
	onlineUsers[u.UserID] = u
}

// RemoveOnlineUser removes a user from the online set.
func RemoveOnlineUser(userID int) {
	delete(onlineUsers, userID)
}

// GetOnlineUser returns the OnlineUser for the given id, or nil.
func GetOnlineUser(userID int) *OnlineUser {
	return onlineUsers[userID]
}

// OnlineUsers returns an iterator over every online user. Callers can
// either drive it with `for u := range game.OnlineUsers()` or collect
// into a slice via slices.Collect.
//
// The iteration order is undefined (Go map order). All consumers must
// run on the main goroutine because the underlying map is not
// synchronised.
func OnlineUsers() iter.Seq[*OnlineUser] {
	return maps.Values(onlineUsers)
}
