package main

import (
	"fmt"
	"net"

	"golang.org/x/net/ipv4"
)

const SACN_PORT = 5568

type SacnClient struct {
	conn      *net.UDPConn
	universes []uint16
}

func NewSacnClient(universes []uint16) (*SacnClient, error) {
	addr := &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: SACN_PORT,
	}
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return nil, err
	}

	for _, universe := range universes {
		multicastAddr := fmt.Sprintf("239.255.%d.%d", universe>>8, universe&0xFF)
		multicastIP := net.ParseIP(multicastAddr)
		p := ipv4.NewPacketConn(conn)

		if err := p.JoinGroup(nil, &net.UDPAddr{IP: multicastIP}); err != nil {
			return nil, err
		}
	}

	return &SacnClient{conn: conn, universes: universes}, nil
}

func (c *SacnClient) Disconnect() error {
	fmt.Println("Disconnecting from lights")

	for _, universe := range c.universes {
		multicastAddr := fmt.Sprintf("239.255.%d.%d", universe>>8, universe&0xFF)
		multicastIP := net.ParseIP(multicastAddr)
		p := ipv4.NewPacketConn(c.conn)
		if err := p.LeaveGroup(nil, &net.UDPAddr{IP: multicastIP}); err != nil {
			return err
		}
	}
	return nil
}

func (c *SacnClient) GetConn() *net.UDPConn {
	return c.conn
}

// func main() {
// 	universes := []uint16{1, 2, 3}
// 	client, err := NewSacnClient(universes)
// 	if err != nil {
// 		fmt.Println("Error creating sACN client:", err)
// 		return
// 	}
// 	defer client.Disconnect()

// 	// Use client.GetConn() to access the UDP connection
// }
