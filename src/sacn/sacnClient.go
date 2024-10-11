package sacn

import (
	"context"
	"fmt"
	"net"
	"sacn_neewer_lite_go/status"

	"github.com/gdamore/tcell/v2"
	"golang.org/x/net/ipv4"
)

const SACN_PORT = 5568

type SacnClient struct {
	conn      *net.UDPConn
	universes []uint16

	status status.Status
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

	return &SacnClient{conn: conn, universes: universes, status: status.NewStatus()}, nil
}

func (c *SacnClient) Disconnect() error {
	c.status.Update("Disconnecting", tcell.ColorYellow)

	for _, universe := range c.universes {
		multicastAddr := fmt.Sprintf("239.255.%d.%d", universe>>8, universe&0xFF)
		multicastIP := net.ParseIP(multicastAddr)
		p := ipv4.NewPacketConn(c.conn)
		if err := p.LeaveGroup(nil, &net.UDPAddr{IP: multicastIP}); err != nil {
			return err
		}
	}
	c.status.Update("Disconnected", tcell.ColorRed)
	return nil
}

func (c *SacnClient) GetConn() *net.UDPConn {
	return c.conn
}

func (c *SacnClient) Listen(ctx context.Context, handler func(*SacnDmxPacket)) {
	c.status.Update("Listening", tcell.ColorGreen)
	buf := make([]byte, 1024)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			n, _, err := c.conn.ReadFromUDP(buf)
			if err != nil {
				c.status.Update("Error reading from UDP connection", tcell.ColorRed)
				return
			}

			if IsDataPacket(buf[:n]) {
				packet, err := SacnPacketFromBytes(buf[:n])

				if err != nil {
					c.status.Update("Error parsing sACN packet", tcell.ColorRed)
				}

				c.status.Increment()
				handler(packet)
			}
		}
	}
}

func (c *SacnClient) GetStatus() *status.Status {
	return &c.status
}
