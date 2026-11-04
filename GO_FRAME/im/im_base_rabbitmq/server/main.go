package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/teslaXvip/go/go_frame/im/im_base_rabbitmq/common"
)

func InitLogger() {
	if err := os.MkdirAll("log", os.ModePerm); err != nil {
		log.Printf("创建日志目录失败: %s", err)
		return
	}
	logFile, err := os.OpenFile("log/im_server.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	if err != nil {
		log.Printf("打开日志文件失败: %s", err)
		return
	}
	log.SetOutput(logFile)
}

var upgrader = websocket.Upgrader{
	HandshakeTimeout: 1 * time.Second,
	ReadBufferSize:   100,
	WriteBufferSize:  100,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func RegistUser(w http.ResponseWriter, r *http.Request) {
	uid := time.Now().UnixMicro() //生成全局唯一
	err := GetRabbitMQ().RegistUser(uid, common.TypeUser)
	if err != nil {
		log.Printf("创建用户%d失败:%s", uid, err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("创建用户失败"))
		return
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(strconv.FormatInt(uid, 10)))
		return
	}
}

func JoinGroup(w http.ResponseWriter, r *http.Request) {
	var gid, uid int64
	var err error
	if gid, err = strconv.ParseInt(r.PathValue("gid"), 10, 64); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("群id非法"))
		return
	}
	if uid, err = strconv.ParseInt(r.PathValue("uid"), 10, 64); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("用户id非法"))
		return
	}
	if err = GetRabbitMQ().AddUser2Group(gid, uid); err != nil {
		log.Printf("用户%d入群%d失败: %s", uid, gid, err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("用户入群失败"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("入群成功"))
}

// 对非法内容进行拦截。比如机器人消息（发言频率过快）；包含欺诈、涉政等违规内容；涉嫌私下联系/交易等。
// 如果需要拦截就返回nil。
func intercept(msg *common.Message) error {
	if len(msg.Content) == 0 {
		return errors.New("内容不能为空")
	}
	return nil
}

// heartBeat 通过ping/pong保持长连接，客户端没有及时响应pong时ReadMessage会超时退出
func heartBeat(conn *websocket.Conn) {
	conn.SetPongHandler(func(appData string) error {
		return nil
	})
	ticker := time.NewTicker(2 * time.Second) // 时间要短于ReadDeadline，确保在超时之前能读到pong
	defer ticker.Stop()
	for range ticker.C {
		conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			return
		}
	}
}

func Speak(w http.ResponseWriter, r *http.Request) {
	// 针对本次请求，创建一个单独的websocket连接
	conn, err := upgrader.Upgrade(w, r, nil) // http升级为websocket协议
	if err != nil {
		log.Printf("upgrade error: %v\n", err)
		return
	}
	// 结束时关闭websocket连接
	defer func() {
		log.Println("close websocket connection")
		conn.Close()
	}()
	go heartBeat(conn) // 心跳保持

	// 接收用户的发言，写入RabbitMQ
	for {
		_, body, err := conn.ReadMessage() // 如果对方主动断开连接或超时，该行会报错，for循环会退出
		if err != nil {
			log.Printf("read error:%v\n", err)
			break
		}
		var msg common.Message
		if err = json.Unmarshal(body, &msg); err != nil {
			log.Printf("json参数解析失败: %s, body: %s", err, body)
			continue
		}
		msg.Content = strings.ReplaceAll(msg.Content, "\n", " ") // 换行符用空格代替。将来要用换行符分隔每条Message，所以一条Message内部不能出现换行符
		msg.Id = time.Now().UnixMicro()                          // 生成全局唯一的消息ID（正规来讲，应该使用SnowFlake等算法）
		msg.Time = time.Now().UnixMicro()                        // 消息的生成时间以服务端接收到消息的时刻为准，防止客户端伪造时间
		if intercept(&msg) != nil {
			continue // 消息被拦截，不能转发给RabbitMQ
		}

		if err = GetRabbitMQ().Send(&msg, msg.To); err != nil {
			log.Printf("用户%s向%s发言失败: %s", msg.From, msg.To, err)
			continue
		}

		if strings.HasPrefix(msg.To, common.TypeUser) {
			// 私聊时，发言人自己这一端也需要看到这条消息（群聊的Exchange里已包含自己的副本）
			if err = GetRabbitMQ().Send(&msg, msg.From); err != nil {
				log.Printf("用户%s发言失败: %s", msg.From, err)
			}
		}
	}
}

func main() {
	InitLogger()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /regist_user", RegistUser)
	mux.HandleFunc("GET /join_group/{gid}/{uid}", JoinGroup)
	mux.HandleFunc("GET /speak", Speak)
	if err := http.ListenAndServe("127.0.0.1:5678", mux); err != nil {
		panic(err)
	}
}
