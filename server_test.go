package ChatRoom

import "testing"

func TestNewServer(t *testing.T) {
	server := NewServer("0.0.0.0", 6666)
	server.Start()
}
