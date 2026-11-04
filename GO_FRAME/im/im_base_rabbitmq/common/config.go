package common

const (
	TypeUser  = "u"
	TypeGroup = "g"
)

const (
	MqUser = "zx"
	MqPass = "zx112233"
)

const (
	GroupMemberPath = "data/im/server/group/" // 群成员信息保存到此目录的文件里
	ReceiveUserPath = "data/im/client/user/"  // 从MQ拉取的消息保存到本地文件中
)
