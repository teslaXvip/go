package model

type User struct {
	Id       int `gorm:"paramaryKey"`
	Name     string
	PassWord string `gorm:"column:password"`
}
