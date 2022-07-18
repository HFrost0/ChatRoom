package main

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

type Server struct {
	IP        string
	Port      int
	OnlineMap map[string]*User
	mapLock   sync.RWMutex
	Message   chan string
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		IP: ip, Port: port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

func (this *Server) Broadcast() {
	for {
		msg := <-this.Message
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

func (this *Server) RemoveUser(user *User) {
	this.mapLock.Lock()
	delete(this.OnlineMap, user.Name)
	this.mapLock.Unlock()
	user.conn.Close()
	fmt.Println(user.conn.RemoteAddr().String(), "下线")
	this.Message <- "[" + user.Addr + "] " + user.Name + " 说：" + "溜了溜了\n"
}

func (this *Server) Handler(conn net.Conn) {
	fmt.Println(conn.RemoteAddr().String(), "建立链接成功")
	user := NewUser(conn)
	defer this.RemoveUser(user)
	go user.ListenMessage() // user对象的chan一旦接收到消息回立刻发送

	this.mapLock.Lock()
	this.OnlineMap[user.Name] = user
	this.mapLock.Unlock()

	this.Message <- "[" + user.Addr + "] " + user.Name + " 说：" + "小逼崽子们，我来了\n" // 广播

	buf := make([]byte, 1024)
	isLive := make(chan bool)
	done := make(chan bool)
	go func() {
		for {
			n, err := conn.Read(buf)
			if err != nil || n == 0 {
				done <- true
				return // 退出监听
			}
			msg := string(buf[:n])
			// 有前缀命令
			if order := strings.Split(msg[:n-1], "|"); len(order) >= 2 {
				switch order[0] {
				case "rename":
					if _, ok := this.OnlineMap[order[1]]; ok {
						user.C <- fmt.Sprintf("用户名[%s]已被使用\n", order[1])
					} else {
						this.Message <- fmt.Sprintf("用户[%s]已更名为[%s]\n", user.Name, order[1])
						this.mapLock.Lock()
						delete(this.OnlineMap, user.Name)
						user.Name = order[1]
						this.OnlineMap[user.Name] = user
						this.mapLock.Unlock()
					}
				case "to":
					if toUser, ok := this.OnlineMap[order[1]]; ok {
						toUser.C <- fmt.Sprintf("[%s] %s 对你说：%s\n", user.Addr, user.Name, order[2])
					} else {
						user.C <- fmt.Sprintf("用户名[%s]不存在\n", order[1])
					}
				}
			} else { // 其他情况广播
				this.Message <- "[" + user.Addr + "] " + user.Name + " 说：" + msg
			}
			isLive <- true
		}
	}()
	for {
		select {
		case <-done:
			return
		case <-isLive:
		case <-time.After(time.Second * 5):
			//user.C <- "由于长时间未活动，你已被强制下线\n"  这样不行，因为已经退出
			user.conn.Write([]byte("由于长时间未活动，你已被强制下线\n"))
			return
		}
	}
}

func (this *Server) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.IP, this.Port))
	defer listener.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("服务器启动成功")

	go this.Broadcast()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go this.Handler(conn)
	}
}
