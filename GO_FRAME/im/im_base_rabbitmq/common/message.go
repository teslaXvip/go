package common

type Message struct {
	Id       int64
	From, To string
	Content  string
	Time     int64
}
