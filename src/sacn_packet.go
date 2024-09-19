package main

import (
	"bytes"
	"encoding/binary"
	"errors"
)

type SacnDmxPacket struct {
	SourceName     string
	Universe       uint16
	Priority       uint8
	SequenceNumber uint8
	Options        uint8
	DmxData        []byte
	Cid            [16]byte
}

func NewSacnDmxPacket(sourceName string, universe uint16, priority, sequenceNumber, options uint8, dmxData []byte, cid [16]byte) *SacnDmxPacket {
	return &SacnDmxPacket{
		SourceName:     sourceName,
		Universe:       universe,
		Priority:       priority,
		SequenceNumber: sequenceNumber,
		Options:        options,
		DmxData:        dmxData,
		Cid:            cid,
	}
}

func FromBytes(packetBytes []byte) (*SacnDmxPacket, error) {
	if len(packetBytes) < 38 {
		return nil, errors.New("byte array too short")
	}

	sourceName := string(packetBytes[44:108])
	sourceName = string(bytes.TrimRight([]byte(sourceName), "\x00"))
	universe := binary.BigEndian.Uint16(packetBytes[113:115])
	priority := packetBytes[108]
	sequenceNumber := packetBytes[111]
	options := packetBytes[112]
	dmxDataLen := int(binary.BigEndian.Uint16(packetBytes[123:125]))
	dmxData := packetBytes[125 : 125+dmxDataLen]
	var cid [16]byte
	copy(cid[:], packetBytes[22:38])

	return &SacnDmxPacket{
		SourceName:     sourceName,
		Universe:       universe,
		Priority:       priority,
		SequenceNumber: sequenceNumber,
		Options:        options,
		DmxData:        dmxData,
		Cid:            cid,
	}, nil
}

func IsDataPacket(packetBytes []byte) bool {
	if len(packetBytes) < 38 {
		return false
	}

	acnPid := packetBytes[4:16]
	if !bytes.Equal(acnPid, []byte("ASC-E1.17\x00\x00\x00")) {
		return false
	}

	vectorRootLayer := binary.BigEndian.Uint32(packetBytes[18:22])
	if vectorRootLayer != 0x00000004 {
		return false
	}

	vectorFramingLayer := binary.BigEndian.Uint32(packetBytes[40:44])
	if vectorFramingLayer != 0x00000002 {
		return false
	}

	vectorDmpLayer := packetBytes[117]
	if vectorDmpLayer != 0x02 {
		return false
	}

	return true
}
