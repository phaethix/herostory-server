package model

// DefaultHeroAvatar is the default avatar assigned to new users.
const DefaultHeroAvatar = "Hero_Shaman"

// User represents a player account, maps to table `t_user`.
type User struct {
	ID            int    `gorm:"column:id;primaryKey;autoIncrement"`
	UserName      string `gorm:"column:user_name;type:varchar(64);uniqueIndex:UK_user_name"`
	Password      string `gorm:"column:password;type:varchar(64)"`
	HeroAvatar    string `gorm:"column:hero_avatar;type:varchar(64)"`
	CreateTime    int64  `gorm:"column:create_time;not null;default:0"`
	LastLoginTime int64  `gorm:"column:last_login_time;not null;default:0"`
}

// TableName returns the table name for gorm.
func (User) TableName() string { return "t_user" }
