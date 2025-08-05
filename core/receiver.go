package core

import (
	"net"
	"sync"
)
// implemented
type HandlerFunc func(data []byte) error
// implemented
type UdpReceiver struct{
	MulticastAddr net.Addr
	conn *net.UDPConn
	bufSize int32
	handler HandlerFunc
	packetBuffer PacketQueue
	mu *sync.Mutex
	cond *sync.Cond

}


func InitUdpReceiver(address string, bufSize int32, handler HandlerFunc) (UdpReceiver, error) {
    addr, err := net.ResolveUDPAddr("udp", address)
    if err != nil {
        return UdpReceiver{}, err
    }

	// try connect multicast
	// WARNING NEED TO CLOSE MANUALLY
	conn, err := net.ListenUDP("udp", addr)
    if err != nil {
        panic(err)
    }

	mu := sync.Mutex{};
    return UdpReceiver{
        MulticastAddr: addr,
		conn: conn,
		bufSize: bufSize,
		handler: handler,
		packetBuffer: *NewPacketQueue(512),
		mu: &mu,
		cond: sync.NewCond(&mu),
    }, nil
}


func (s *UdpReceiver) Close(){
	s.conn.Close();
}
func (s *UdpReceiver) Listen() error {
	// Consumer goroutine
	go func() {
		for {
			s.mu.Lock()
			for s.packetBuffer.Length() == 0 {
				s.cond.Wait()
			}
			s.mu.Unlock()

			data, _ := s.packetBuffer.Pop()
			s.handler(data)
		}
	}()

	// Producer loop
	for {
		buf := make([]byte, s.bufSize)
		_, _, err := s.conn.ReadFromUDP(buf)
		if err != nil {
			return err
		}
		s.packetBuffer.Append(buf)

		s.mu.Lock()
		s.cond.Signal() // Wake consumer if it's waiting
		s.mu.Unlock()
	}
}
