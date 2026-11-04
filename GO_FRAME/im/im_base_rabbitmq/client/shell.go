package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/desertbit/grumble"
	"github.com/nsf/termbox-go"
	"github.com/teslaXvip/go/go_frame/im/im_base_rabbitmq/common"
)

var (
	app *grumble.App

	waterfall = make(chan struct{}, 1) // 是否要终止消息瀑布流的展示
	ugReg     *regexp.Regexp           // 判断msg.To是否合法
)

func init() {
	ugReg = regexp.MustCompile(`^[ug]\d+$`)
}

func listenForKeyPress() {
	for {
		ev := termbox.PollEvent()
		switch ev.Type {
		case termbox.EventKey:
			if ev.Key == termbox.KeyCtrlA { // 按下Ctrl+A键时终止消息瀑布流
				select {
				case waterfall <- struct{}{}:
				default: // 不在瀑布流中，忽略
				}
			}
		case termbox.EventError:
			log.Printf("termbox event error: %s", ev.Err)
			return
		}
	}
}

func InitApp() {
	app = grumble.New(&grumble.Config{
		Name:        "dqq-im",
		Description: "基于RabbitMQ的聊天系统",
	})

	app.OnInit(func(a *grumble.App, flags grumble.FlagMap) error {
		if err := termbox.Init(); err != nil {
			fmt.Println("Failed to initialize termbox:", err)
			os.Exit(1)
		}
		go listenForKeyPress()
		return nil
	})

	app.AddCommand(&grumble.Command{
		Name: "regist",
		Help: "注册新用户",
		Run: func(c *grumble.Context) error {
			uid := RegistUser()
			if uid <= 0 {
				c.App.Println("注册失败")
			} else {
				c.App.Println("注册成功，用户id是", uid)
			}
			return nil
		},
	})

	app.AddCommand(&grumble.Command{
		Name: "join",
		Help: "加入群组",
		Flags: func(f *grumble.Flags) { // join -u 3 -g 1744695500829846
			f.Int64("u", "uid", 0, "用户id")
			f.Int64("g", "gid", 0, "群组id")
		},
		Run: func(c *grumble.Context) error {
			gid, uid := c.Flags.Int64("gid"), c.Flags.Int64("uid")
			if gid <= 0 || uid <= 0 {
				c.App.Println("请用 -u 和 -g 指定用户id与群组id")
				return nil
			}
			if JoinGroup(gid, uid) {
				c.App.Println("加入群组成功")
			} else {
				c.App.Println("加入群组失败")
			}
			return nil
		},
	})

	app.AddCommand(&grumble.Command{
		Name: "login",
		Help: "登录",
		Args: func(a *grumble.Args) { // login 1
			a.Int64("uid", "用户id")
		},
		Run: func(c *grumble.Context) error {
			uid := c.Args.Int64("uid")
			if uid <= 0 {
				c.App.Println("请提供合法的用户id")
				return nil
			}
			rabbit, err := GetRabbitMQ()
			if err != nil {
				log.Printf("login failed: %s", err)
				c.App.Println("登录失败：无法连接RabbitMQ")
				return nil
			}
			setWindowFrom(common.TypeUser + strconv.FormatInt(uid, 10)) // 更新当前登录者
			switchWindow("")                                            // 清空会话窗口
			rabbit.Receive(uid)                                         // 开始消费自己的队列
			c.App.Println("登录成功，当前用户是", uid)
			return nil
		},
	})

	app.AddCommand(&grumble.Command{
		Name: "send",
		Help: "发送消息",
		Flags: func(f *grumble.Flags) {
			f.String("t", "to", "", "消息发送给谁，发给个人请用u+uid，发给群请使用g+群id")
			f.String("m", "message", "", "消息内容")
		},
		Run: func(c *grumble.Context) error {
			from := currentFrom()
			// 流程检查
			if len(from) == 0 {
				c.App.Println("请先登录")
				return nil
			}
			// 参数合法性检查
			to := c.Flags.String("to")
			if len(to) == 0 {
				c.App.Println("请指定to")
				c.App.Println()
				return nil
			}
			if !ugReg.MatchString(to) {
				c.App.Println("请输入合理的用户或群组标识")
				return nil
			}
			content := c.Flags.String("message")
			if len(content) == 0 {
				c.App.Println("请指定message")
				c.App.Println()
				return nil
			}

			// 聊天对象发生了变化（即切换了聊天窗口）
			if _, curTo := windowSnapshot(); curTo != to {
				switchWindow(to)
			}

			// 把消息发给web server，由server投递到RabbitMQ
			msg := &common.Message{
				Id:      time.Now().UnixNano(),
				From:    from,
				To:      to,
				Content: content,
				Time:    time.Now().UnixMicro(),
			}
			if err := Send(msg); err != nil {
				c.App.Println("消息发送失败")
				c.App.Println()
				return nil
			}

			// 展示消息瀑布流
			dup := make(Set[int64], 1000) // 按消息ID排重
			// 1. 先从文件里读取历史消息展示给用户
			printHistory(c, from, to, dup)
			// 2. 再展示channel里的实时消息，直到用户按下Ctrl+A键时退出
			drainWaterfall()
		LOOP:
			for {
				ch, _ := windowSnapshot()
				select {
				case ele, ok := <-ch: // 这部分消息不能保证是按时间排序的
					if !ok { // 会话窗口已切换，下一轮会读到新的channel
						continue
					}
					var m common.Message
					if err := json.Unmarshal(ele, &m); err != nil {
						log.Printf("Unmarshal line failed: %s", err)
						continue
					}
					if _, exists := dup[m.Id]; !exists { // 按消息ID排重
						dup[m.Id] = struct{}{}
						printMsg(c, &m)
					}
				case <-waterfall:
					break LOOP
				}
			}
			return nil
		},
	})
}

// printHistory 从本地文件里读取当前会话的历史消息，展示最近10条
func printHistory(c *grumble.Context, from, to string, dup Set[int64]) {
	path := filepath.Join(common.ReceiveUserPath+from, to)
	bs, err := os.ReadFile(path)
	if err != nil {
		return // 还没有历史消息
	}
	msgs := make([]*common.Message, 0, 100)
	for _, line := range strings.Split(string(bs), "\n") {
		if len(line) == 0 {
			continue
		}
		var msg common.Message
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			log.Printf("Unmarshal line failed: %s", err)
			continue
		}
		if _, exists := dup[msg.Id]; exists {
			continue
		}
		dup[msg.Id] = struct{}{}
		msgs = append(msgs, &msg)
	}
	sort.Slice(msgs, func(i, j int) bool {
		return msgs[i].Time < msgs[j].Time // 按时间顺序排序
	})
	if len(msgs) > 10 { // 只截取最后的10条进行展示
		msgs = msgs[len(msgs)-10:]
	}
	for _, msg := range msgs {
		printMsg(c, msg)
	}
}

func printMsg(c *grumble.Context, msg *common.Message) {
	show := fmt.Sprintf("%s  %s: %s",
		time.Unix(msg.Time/1e6, 1e3*msg.Time%1e6).Format("2006-01-02 15:04:05.000000"),
		msg.From, msg.Content)
	c.App.Println(show)
}

// drainWaterfall 清掉残留的终止信号，避免刚进入瀑布流就退出
func drainWaterfall() {
	select {
	case <-waterfall:
	default:
	}
}

func CloseShell() {
	if app != nil {
		app.Close()
	}
	termbox.Close()
}
