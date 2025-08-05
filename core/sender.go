package core

import (
	"net"
)

type  UdpSender struct{
	MulticastAddr net.Addr
	conn *net.UDPConn
    bufSize int32
}

func InitUdpSender(address string, bufSize int32) (UdpSender, error) {
    addr, err := net.ResolveUDPAddr("udp", address)
    if err != nil {
        return UdpSender{}, err
    }

	// WARNING NEED TO CLOSE MANUALLY
	conn, err := net.DialUDP("udp", nil, addr)
    if err != nil {
        panic(err)
    }

    return UdpSender{
        MulticastAddr: addr,
		conn: conn,
        bufSize: bufSize,
    }, nil
}

// Closing connection at here
func (s *UdpSender) Close(){
	s.conn.Close();
}
func (s *UdpSender) Send(msg []byte ) error{
    _, err := s.conn.Write([]byte(msg))
    if err != nil {
		return err
	}
	return nil
}