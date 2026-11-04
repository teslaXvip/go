package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/teslaXvip/go/go_frame/im/im_base_rabbitmq/common"
)

// TestMain 在所有用例结束后关闭全局连接，避免测试进程残留连接
func TestMain(m *testing.M) {
	code := m.Run()
	if mq != nil {
		if mq.mqChannel != nil {
			_ = mq.mqChannel.Close()
		}
		if mq.mqConn != nil {
			_ = mq.mqConn.Close()
		}
	}
	os.Exit(code)
}

// dialTestMQ 预检 RabbitMQ 是否可用，不可用时跳过用例，避免 GetRabbitMQ 内部 panic
func dialTestMQ(t *testing.T) *amqp.Connection {
	t.Helper()
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@localhost:5672/", common.MqUser, common.MqPass))
	if err != nil {
		t.Skipf("RabbitMQ 不可用，跳过测试: %v", err)
	}
	return conn
}

// uniqueID 生成用例内唯一的 uid/gid，避免与已有持久化队列冲突
func uniqueID() int64 {
	return time.Now().UnixNano()
}

func TestGetRabbitMQ(t *testing.T) {
	conn := dialTestMQ(t)
	defer conn.Close()

	first := GetRabbitMQ()
	second := GetRabbitMQ()
	if first == nil || second == nil {
		t.Fatal("GetRabbitMQ 返回 nil")
	}
	if first != second {
		t.Fatal("GetRabbitMQ 未返回同一实例")
	}
	if first.mqConn == nil || first.mqChannel == nil {
		t.Fatal("RabbitMQ 连接或 Channel 未初始化")
	}
}

func TestRegistUser(t *testing.T) {
	conn := dialTestMQ(t)
	defer conn.Close()

	uid := uniqueID()
	user := common.TypeUser + strconv.FormatInt(uid, 10)
	queues := []string{user + "_computer", user + "_mobile"}

	ch, err := conn.Channel()
	if err != nil {
		t.Fatalf("创建校验 Channel 失败: %v", err)
	}
	t.Cleanup(func() {
		for _, queue := range queues {
			_, _ = ch.QueueDelete(queue, false, false, false)
		}
		_ = ch.ExchangeDelete(user, false, false)
		_ = ch.Close()
	})

	// 注册用户：声明 Exchange、声明两个队列并绑定
	if err := GetRabbitMQ().RegistUser(uid, common.TypeUser); err != nil {
		t.Fatalf("RegistUser 失败: %v", err)
	}
	// 重复注册应保持幂等
	if err := GetRabbitMQ().RegistUser(uid, common.TypeUser); err != nil {
		t.Fatalf("重复 RegistUser 失败: %v", err)
	}

	// Exchange 应为可用的 fanout 类型
	if err := ch.ExchangeDeclarePassive(user, "fanout", true, false, false, false, nil); err != nil {
		t.Fatalf("Exchange %s 未按预期声明: %v", user, err)
	}
	// 两个队列都应存在
	for _, queue := range queues {
		if _, err := ch.QueueInspect(queue); err != nil {
			t.Errorf("队列 %s 未按预期声明: %v", queue, err)
		}
	}

	// 向 Exchange 投递消息，两个队列都应能收到，验证绑定关系
	body := []byte("hello " + user)
	if err := ch.Publish(user, "", false, false, amqp.Publishing{Body: body}); err != nil {
		t.Fatalf("向 Exchange %s 投递消息失败: %v", user, err)
	}
	for _, queue := range queues {
		msg, ok, err := ch.Get(queue, true)
		if err != nil {
			t.Errorf("从队列 %s 读取消息失败: %v", queue, err)
			continue
		}
		if !ok {
			t.Errorf("队列 %s 未收到 Exchange 投递的消息，绑定可能失效", queue)
			continue
		}
		if string(msg.Body) != string(body) {
			t.Errorf("队列 %s 收到的消息不匹配: got %q, want %q", queue, msg.Body, body)
		}
	}
}

func TestAddUser2Group(t *testing.T) {
	conn := dialTestMQ(t)
	defer conn.Close()

	gid := uniqueID()
	uid1, uid2 := uniqueID(), uniqueID()+1
	group := common.TypeGroup + strconv.FormatInt(gid, 10)
	memberFile := filepath.Join(common.GroupMemberPath, strconv.FormatInt(gid, 10))

	var queues []string
	for _, uid := range []int64{uid1, uid2} {
		user := group + "_" + common.TypeUser + strconv.FormatInt(uid, 10)
		queues = append(queues, user+"_computer", user+"_mobile")
	}

	ch, err := conn.Channel()
	if err != nil {
		t.Fatalf("创建校验 Channel 失败: %v", err)
	}
	t.Cleanup(func() {
		for _, queue := range queues {
			_, _ = ch.QueueDelete(queue, false, false, false)
		}
		_ = ch.ExchangeDelete(group, false, false)
		_ = ch.Close()
		_ = os.Remove(memberFile)
	})

	// 将两个用户加入群
	if err := GetRabbitMQ().AddUser2Group(gid, uid1, uid2); err != nil {
		t.Fatalf("AddUser2Group 失败: %v", err)
	}

	// 群成员应写入本地文件
	content, err := os.ReadFile(memberFile)
	if err != nil {
		t.Fatalf("读取群成员文件 %s 失败: %v", memberFile, err)
	}
	for _, uid := range []int64{uid1, uid2} {
		if !strings.Contains(string(content), strconv.FormatInt(uid, 10)+"\n") {
			t.Errorf("群成员文件缺少用户 %d，实际内容: %q", uid, content)
		}
	}

	// 群 Exchange 应为可用的 fanout 类型
	if err := ch.ExchangeDeclarePassive(group, "fanout", true, false, false, false, nil); err != nil {
		t.Fatalf("Exchange %s 未按预期声明: %v", group, err)
	}
	for _, queue := range queues {
		if _, err := ch.QueueInspect(queue); err != nil {
			t.Errorf("队列 %s 未按预期声明: %v", queue, err)
		}
	}

	// 向群 Exchange 投递消息，每个成员的每个设备队列都应收到
	body := []byte("hello " + group)
	if err := ch.Publish(group, "", false, false, amqp.Publishing{Body: body}); err != nil {
		t.Fatalf("向 Exchange %s 投递消息失败: %v", group, err)
	}
	for _, queue := range queues {
		msg, ok, err := ch.Get(queue, true)
		if err != nil {
			t.Errorf("从队列 %s 读取消息失败: %v", queue, err)
			continue
		}
		if !ok {
			t.Errorf("队列 %s 未收到群消息，绑定可能失效", queue)
			continue
		}
		if string(msg.Body) != string(body) {
			t.Errorf("队列 %s 收到的消息不匹配: got %q, want %q", queue, msg.Body, body)
		}
	}
}
