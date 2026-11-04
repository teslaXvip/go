package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/teslaXvip/go/go_frame/im/im_base_rabbitmq/common"
)

type RabbitMQ struct {
	mqConn    *amqp.Connection
	mqChannel *amqp.Channel // 支持并发使用
}

// Set 简化的集合
type Set[T comparable] map[T]struct{}

var (
	mq     *RabbitMQ
	mqOnce sync.Once
	mqErr  error

	windowLck  sync.RWMutex
	windowFrom string                   // 当前登录者，形如u123
	windowTo   string                   // 当前会话窗口的聊天对象，形如u456或g789
	windowChan = make(chan []byte, 100) // 仅属于当前会话窗口的消息

	groupBelongLck sync.RWMutex
	groupBelong    map[string]Set[string] // uid -> 该用户所在的gid集合

	consumedLck sync.Mutex
	consumed    = make(map[string]struct{}) // 已启动消费的队列，避免重复消费
)

func GetRabbitMQ() (*RabbitMQ, error) {
	mqOnce.Do(func() {
		// 连接RabbitMQ
		conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@localhost:5672/", common.MqUser, common.MqPass))
		if err != nil {
			mqErr = fmt.Errorf("connect to RabbitMQ failed: %w", err)
			return
		}
		// 基于连接创建Channel
		ch, err := conn.Channel()
		if err != nil {
			_ = conn.Close()
			mqErr = fmt.Errorf("open channel failed: %w", err)
			return
		}
		mq = &RabbitMQ{
			mqConn:    conn,
			mqChannel: ch,
		}
	})
	return mq, mqErr
}

// InitGroupBelong 填充groupBelong：服务端把群成员写到了GroupMemberPath下的文件里，文件名是gid，内容是uid
func InitGroupBelong() {
	belong := make(map[string]Set[string], 10)
	entries, err := os.ReadDir(common.GroupMemberPath)
	if err != nil {
		log.Printf("读取群成员目录%s失败: %s", common.GroupMemberPath, err)
	} else {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			gid := entry.Name()
			if _, err := strconv.ParseInt(gid, 10, 64); err != nil {
				continue // 文件名不是群id，跳过
			}
			bs, err := os.ReadFile(filepath.Join(common.GroupMemberPath, gid))
			if err != nil {
				log.Printf("读取群成员文件%s失败: %s", gid, err)
				continue
			}
			for _, line := range strings.Split(string(bs), "\n") {
				uid := strings.TrimSpace(line)
				if len(uid) == 0 {
					continue
				}
				if belong[uid] == nil {
					belong[uid] = make(Set[string], 10)
				}
				belong[uid][gid] = struct{}{}
			}
		}
	}

	groupBelongLck.Lock()
	groupBelong = belong
	groupBelongLck.Unlock()
}

// Receive 从MQ接收发给uid的消息及其所在群的群消息，写入本地文件，当前会话窗口的消息同时写入channel
func (mq *RabbitMQ) Receive(uid int64) {
	self := common.TypeUser + strconv.FormatInt(uid, 10)
	// 一对一发给uid的消息全在这个队列里
	userPath := common.ReceiveUserPath + self
	if err := os.MkdirAll(userPath, os.ModePerm); err != nil {
		log.Printf("创建目录%s失败: %s", userPath, err)
		return
	}
	mq.readQueue(self, common.TypeUser, userPath, self)

	// uid在这些群里面，群里有消息时也需要写入本地文件
	groupBelongLck.RLock()
	gids := groupBelong[strconv.FormatInt(uid, 10)]
	groupBelongLck.RUnlock()
	for gid := range gids {
		// 发到群gid的消息全在这个队列里，每个群成员都拥有一份副本
		queue := common.TypeGroup + gid + "_" + self
		mq.readQueue(queue, common.TypeGroup, userPath, self)
	}
}

// readQueue 把MQ指定队列里的消息拉下来，存入本地文件，并挑出当前会话窗口的消息写入channel
func (mq *RabbitMQ) readQueue(queue, qtype, path, self string) {
	consumedLck.Lock()
	if _, exists := consumed[queue]; exists {
		consumedLck.Unlock()
		return
	}
	deliveryCh, err := mq.mqChannel.Consume(
		queue+"_computer", // queue，PC端聊天
		"",                // consumer
		false,             // auto-ack，成功写入本地文件后再手动ACK
		false,             // exclusive
		false,             // no-local
		false,             // no-wait
		nil,               // args
	)
	if err != nil {
		consumedLck.Unlock()
		log.Printf("消费队列%s失败: %s", queue+"_computer", err)
		return
	}
	consumed[queue] = struct{}{}
	consumedLck.Unlock()

	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Println(err)
			}
		}()
		log.Printf("开始消费队列%s", queue+"_computer")
		for delivery := range deliveryCh { // 读取MQ里的消息
			var msg common.Message
			if err := json.Unmarshal(delivery.Body, &msg); err != nil {
				log.Printf("Unmarshal message failed: %s", err)
				_ = delivery.Nack(false, false)
				continue
			}
			// 一个会话对应一个本地文件：群聊按群归档，私聊按对方归档
			session := msg.From
			if qtype == common.TypeGroup {
				session = msg.To
			} else if session == self {
				session = msg.To
			}
			if session == "" {
				_ = delivery.Nack(false, false)
				continue
			}
			appendSessionMessage(path, session, delivery.Body)
			_ = delivery.Ack(false) // 成功写入本地文件后，再ACK，标记消息消费成功

			if matchWindow(queue, msg) {
				pushWindow(delivery.Body)
			}
		}
	}()
}

// appendSessionMessage 把消息追加到本地会话文件path/session
func appendSessionMessage(path, session string, body []byte) {
	file := filepath.Join(path, session)
	fout, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		log.Printf("打开文件%s失败: %s", file, err)
		return
	}
	defer fout.Close()
	if _, err := fout.Write(append(body, '\n')); err != nil {
		log.Printf("写入文件%s失败: %s", file, err)
	}
}

// matchWindow 判断消息是否属于当前会话窗口
func matchWindow(queue string, msg common.Message) bool {
	windowLck.RLock()
	to := windowTo
	windowLck.RUnlock()

	if queue == msg.From { // 用户自己发出去的消息
		return true
	}
	if strings.HasPrefix(to, common.TypeUser) && to == msg.From { // 对方发给当前用户的
		return true
	}
	if strings.HasPrefix(to, common.TypeGroup) && to == msg.To { // 发到这个群里面的
		return true
	}
	return false
}

// pushWindow 把消息写入当前会话窗口所在的channel，channel满了就丢弃，避免阻塞消费
func pushWindow(body []byte) {
	windowLck.RLock()
	defer windowLck.RUnlock()
	select {
	case windowChan <- body:
	default:
	}
}

// switchWindow 切换会话窗口：换一个新的channel，并更新windowTo
func switchWindow(to string) {
	windowLck.Lock()
	close(windowChan)
	windowChan = make(chan []byte, 100)
	windowTo = to
	windowLck.Unlock()
}

func windowSnapshot() (chan []byte, string) {
	windowLck.RLock()
	defer windowLck.RUnlock()
	return windowChan, windowTo
}

func setWindowFrom(from string) {
	windowLck.Lock()
	windowFrom = from
	windowLck.Unlock()
}

func currentFrom() string {
	windowLck.RLock()
	defer windowLck.RUnlock()
	return windowFrom
}

// Release 释放MQ连接
func (mq *RabbitMQ) Release() {
	if mq == nil {
		return
	}
	if mq.mqChannel != nil {
		_ = mq.mqChannel.Close()
	}
	if mq.mqConn != nil {
		_ = mq.mqConn.Close()
	}
}
