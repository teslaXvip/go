package model

import "time"

type News struct {
	Id           int        `gorm:"primaryKey"`
	UserId       int        //发布者id
	UserName     string     `gorm:"-"` //数据库里没有这一列
	Title        string     //新闻标题
	Content      string     `gorm:"column:article"`     //正文
	PostTime     *time.Time `gorm:"column:create_time"` //发布时间
	DeleteTime   *time.Time //删除时间
	ViewPostTime string     `gorm:"-"`
}
