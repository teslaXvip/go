package model

type User struct {
	Id       int    `gorm:"column:id;primaryKey;autoIncrement"`
	Name     string `gorm:"column:name;not null;uniqueIndex:idx_user_name"`
	PassWord string `gorm:"column:password;not null"`
}

func (User) TableName() string {
	return "user"
}
