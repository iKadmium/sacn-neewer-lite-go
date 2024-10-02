package sacn

import (
	"fmt"
	"net"
	"time"

	"golang.org/x/net/ipv4"
)

const SACN_PORT = 5568

type SacnClient struct {
	conn          *net.UDPConn
	universes     []uint16
	packetCount   uint16
	lastPrintTime time.Time
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

	return &SacnClient{conn: conn, universes: universes, packetCount: 0, lastPrintTime: time.Now()}, nil
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

func (c *SacnClient) Listen(handler func(*SacnDmxPacket)) {
	buf := make([]byte, 1024)

	go func() {
		timer := time.NewTimer(time.Second)
		for range timer.C {
			fmt.Printf("sACN packets received in the last second: %d\n", c.packetCount)
			c.packetCount = 0
			c.lastPrintTime = time.Now()
			timer.Reset(time.Second)
		}
	}()

	for {
		n, _, err := c.conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error reading from UDP connection:", err)
			return
		}

		if IsDataPacket(buf[:n]) {
			packet, err := SacnPacketFromBytes(buf[:n])

			if err != nil {
				fmt.Println("Error parsing sACN packet:", err)
				continue
			}

			handler(packet)

			// Packet counting logic
			c.packetCount++
		}

	}
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
