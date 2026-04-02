package game

// OnlineUser holds the runtime data of a logged-in user.
// All functions operating on onlineUsers are called from the main thread, no lock needed.
type OnlineUser struct {
	UserID     int
	UserName   string
	HeroAvatar string
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

// ForEachOnlineUser iterates over all online users and calls fn for each.
func ForEachOnlineUser(fn func(u *OnlineUser)) {
	for _, u := range onlineUsers {
		fn(u)
	}
}
