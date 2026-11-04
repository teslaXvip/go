package main

type Hub struct {
	// 注册客户端
	register chan *Client
	// 注销客户端
	unregister chan *Client
	// 广播消息
	broadcast chan []byte
	// 保存所有在线客户端
	clients map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
		clients:    make(map[*Client]struct{}),
	}
}

func (hub *Hub) Run() {
	//通过select机制避免了并发修改了clients; 避免了遍历clients的同时修改clients。
	for {
		select {
		case client := <-hub.register:
			hub.clients[client] = struct{}{} //注册client
		case client := <-hub.unregister:
			if _, ok := hub.clients[client]; ok { //防止重复注销
				delete(hub.clients, client) //注销client
				close(client.send)          //hub从此以后不需要再向该client广播消息了
			}
		case msg := <-hub.broadcast:
			for client := range hub.clients {
				client.send <- msg
			}
		}
	}
}
