package ChatRoom

import "net"

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn
}

func NewUser(conn net.Conn) *User {
	user := &User{
		Name: conn.RemoteAddr().String(),
		Addr: conn.RemoteAddr().String(),
		C:    make(chan string),
		conn: conn,
	}
	return user
}

func (user *User) ListenMessage() {
	for {
		msg := <-user.C
		user.conn.Write([]byte(msg))
	}
}
