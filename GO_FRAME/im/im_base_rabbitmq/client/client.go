package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/teslaXvip/go/go_frame/im/im_base_rabbitmq/common"
)

var (
	host   = "127.0.0.1:5678"
	wsConn *websocket.Conn
)

func RegistUser() (uid int64) {
	resp, err := http.Get("http://" + host + "/regist_user")
	if err != nil {
		log.Printf("regist user failed: %s", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("resp code %d", resp.StatusCode)
		return
	}
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("read resp body failed: %s", err)
		return
	}
	uid, err = strconv.ParseInt(string(bs), 10, 64)
	if err != nil {
		log.Printf("invalid uid [%s]", string(bs))
		return 0
	}
	return
}

func JoinGroup(gid, uid int64) (success bool) {
	resp, err := http.Get(fmt.Sprintf("http://%s/join_group/%d/%d", host, gid, uid))
	if err != nil {
		log.Printf("join group failed: %s", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		success = true
	} else {
		log.Printf("resp code %d", resp.StatusCode)
	}
	return
}

// Send 把消息发给web server，由server投递到RabbitMQ
func Send(msg *common.Message) error {
	if wsConn == nil {
		conn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://%s/speak", host), nil)
		if err != nil {
			log.Printf("connect to websocket server failed: %s", err)
			return err
		}
		// 收到ping后，默认情况就会自动返回pong
		conn.SetPingHandler(func(appData string) error {
			return conn.WriteMessage(websocket.PongMessage, nil)
		})
		// 收到CloseMessage后，默认行为就是关闭连接
		conn.SetCloseHandler(func(code int, text string) error {
			return conn.Close()
		})
		// 持续ReadMessage才能激活PingHandler、CloseHandler，并感知连接断开
		go func() {
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					return
				}
			}
		}()
		wsConn = conn
	}

	if err := wsConn.WriteJSON(msg); err != nil {
		log.Printf("send message failed: %s", err)
		return err
	}
	return nil
}

func ReleaseWebSocket() {
	if wsConn != nil {
		_ = wsConn.WriteMessage(websocket.CloseMessage, nil)
		_ = wsConn.Close()
		wsConn = nil
	}
}
