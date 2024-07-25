package am2manager

import (
	"bytes"
	"crypto/sha256"
	"fmt"
)

const (
	am2dataLength = 6204
	am2Length     = 6144
)

var InitData = [...]byte{0, 0, 50, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 100, 100, 60, 30, 0, 0, 0, 0, 200, 66, 0, 0, 192, 192, 0, 0, 192, 64, 0, 0, 250, 68, 0, 0, 192, 192, 0, 0, 192, 64, 0, 0, 122, 69, 0, 0, 192, 192, 0, 0, 192, 64}

type Am2Data struct {
	Level      byte
	GainMin    byte
	GainMax    byte
	Mix        byte
	Am2        []byte
	OriginData []byte
}

func (d Am2Data) HashAm2() string {
	hash := sha256.Sum256(d.Am2)
	return fmt.Sprintf("%x", hash[:])
}

func (d Am2Data) HashData() string {
	data, _ := d.MarshalBinary()
	if len(data) == 0 {
		return ""
	}
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash[:])
}

func (d Am2Data) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	if len(d.OriginData) == 60 {
		buf.Write(d.OriginData[:])
	} else {
		buf.Write(InitData[:])
	}
	buf.Write(d.Am2)
	cp := buf.Bytes()
	cp[0x12] = d.Mix
	cp[0x13] = d.Level
	cp[0x14] = d.GainMax
	cp[0x15] = d.GainMin
	return cp, nil
}

func (d Am2Data) String() string {
	return fmt.Sprintf("AM2DATA: Level = %d Mix = %d\nGainMin = %d GainMax = %d", d.Level, d.Mix, d.GainMin, d.GainMax)
}

func (d *Am2Data) UnmarshalBinary(data []byte) error {
	if d == nil {
		return fmt.Errorf("am2data can't be nil")
	}

	switch {
	case IsAm2Data(data):
		d.OriginData = data[0:60]
		d.Am2 = data[60:]
		d.Mix = data[0x12]
		d.Level = data[0x13]
		d.GainMax = data[0x14]
		d.GainMin = data[0x15]
	case IsAm2(data):
		d.Am2 = data
		d.Mix = 100
		d.Level = 100
		d.GainMax = 60
		d.GainMin = 30
	default:
		return fmt.Errorf("invalid data type")

	}
	return nil
}

func IsAm2(data []byte) bool {
	return len(data) == am2Length
}

func IsAm2Data(data []byte) bool {
	return len(data) == am2dataLength
}
