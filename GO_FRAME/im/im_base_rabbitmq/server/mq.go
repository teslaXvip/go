package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/teslaXvip/go/go_frame/im/im_base_rabbitmq/common"
)

type RabbitMQ struct {
	mqConn    *amqp.Connection
	mqChannel *amqp.Channel // 支持并发使用
}

var (
	mq     *RabbitMQ
	mqOnce sync.Once
)

func GetRabbitMQ() *RabbitMQ {
	mqOnce.Do(func() {
		// 连接RabbitMQ
		conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@localhost:5672/", common.MqUser, common.MqPass))
		if err != nil {
			log.Panicf("connect to RabbitMQ failed: %s", err)
		}
		// 基于连接创建Channel
		ch, err := conn.Channel()
		if err != nil {
			log.Panicf("open channel failed: %s", err)
		}
		mq = &RabbitMQ{
			mqConn:    conn,
			mqChannel: ch,
		}
	})
	return mq
}

// 用户注册
func (mq *RabbitMQ) RegistUser(uid int64, userType string) error {
	user := userType + strconv.FormatInt(uid, 10)
	exchange := user // 以用户ID来命名Exchange
	// 声明Exchange
	err := mq.mqChannel.ExchangeDeclare(
		exchange,
		"fanout", // type
		true,     // durable
		false,    // auto delete
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		return err
	}

	// 一个用户对应两个队列
	queues := []string{user + "_computer", user + "_mobile"}
	for _, QueueName := range queues {
		// 声明队列
		_, err := mq.mqChannel.QueueDeclare(
			QueueName,
			true,  // durable
			false, // delete when unused
			false, // exclusive
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			return err
		}
		// Exchange和Queue绑定
		err = mq.mqChannel.QueueBind(
			QueueName, // Queue Name
			"",        // routing key, fanout模式下会忽略routing key
			exchange,
			false, // nowait
			nil,   // arguments
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// Send 把消息投递到指定的Exchange（用户的u<uid>或群的g<gid>）
func (mq *RabbitMQ) Send(message *common.Message, exchange string) error {
	bytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// 指定超时
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 向MQ发送消息
	err = mq.mqChannel.PublishWithContext(
		ctx,
		exchange,
		"",    // routing key，fanout模式下会忽略routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json", //MIME content type
			Body:         bytes,
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func (mq *RabbitMQ) AddUser2Group(gid int64, uids ...int64) error {
	// 把群成员信息写入本地文件（正规来讲，应该写入数据库）
	os.MkdirAll(common.GroupMemberPath, os.ModePerm)
	fout, err := os.OpenFile(common.GroupMemberPath+strconv.FormatInt(gid, 10), os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	for _, uid := range uids {
		fout.WriteString(strconv.FormatInt(uid, 10) + "\n")
	}
	fout.Close()

	// 创建群对应的Queue和Exchange，及其绑定关系
	group := common.TypeGroup + strconv.FormatInt(gid, 10)
	exchange := group // 以群ID来命名Exchange
	// 声明Exchange
	err = mq.mqChannel.ExchangeDeclare(
		exchange,
		"fanout", // type
		true,     // durable
		false,    // auto delete
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		return err
	}

	// 声明队列
	for _, uid := range uids {
		user := group + "_" + common.TypeUser + strconv.FormatInt(uid, 10)
		// 一个用户对应两个队列，g_u_computer和g_u_mobile
		queues := []string{user + "_computer", user + "_mobile"}
		for _, QueueName := range queues {
			_, err := mq.mqChannel.QueueDeclare(
				QueueName,
				true,  // durable
				false, // delete when unused
				false, // exclusive
				false, // no-wait
				nil,   // arguments
			)
			if err != nil {
				return err
			}
			// Exchange和Queue绑定
			err = mq.mqChannel.QueueBind(
				QueueName, // Queue Name
				"",        // routing key, fanout模式下会忽略routing key
				exchange,
				false, // nowait
				nil,   // arguments
			)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
