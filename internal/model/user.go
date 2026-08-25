package model

// DefaultHeroAvatar is assigned to newly registered accounts.
const DefaultHeroAvatar = "Hero_Shaman"

// DefaultMaxHp is the HP a new account and a fresh login session start with.
const DefaultMaxHp int32 = 100

// User maps to table t_user.
type User struct {
	ID            int    `gorm:"column:id;primaryKey;autoIncrement"`
	UserName      string `gorm:"column:user_name;type:varchar(64);uniqueIndex:UK_user_name"`
	Password      string `gorm:"column:password;type:varchar(64)"`
	HeroAvatar    string `gorm:"column:hero_avatar;type:varchar(64)"`
	CurrHp        int32  `gorm:"column:curr_hp;not null;default:100"`
	CreateTime    int64  `gorm:"column:create_time;not null;default:0"`
	LastLoginTime int64  `gorm:"column:last_login_time;not null;default:0"`
}

func (User) TableName() string { return "t_user" }
