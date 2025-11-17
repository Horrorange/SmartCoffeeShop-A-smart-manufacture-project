package gateway

import (
	"errors"
	"time"

	modbus "github.com/goburrow/modbus"
)

type Grinder struct {
	Host string
	Port int
}

func (g *Grinder) addr() string {
	return g.Host + ":" + itoa(g.Port)
}

func (g *Grinder) client() (modbus.Client, func(), error) {
	h := modbus.NewTCPClientHandler(g.addr())
	h.Timeout = 5 * time.Second
	if err := h.Connect(); err != nil {
		return nil, func() {}, err
	}
	c := modbus.NewClient(h)
	return c, func() { h.Close() }, nil
}

const (
	cmdReg       = 0
	statusReg    = 1
	beanLevelReg = 2
)

func (g *Grinder) readU16(c modbus.Client, addr uint16) (uint16, error) {
	b, err := c.ReadHoldingRegisters(addr, 1)
	if err != nil {
		return 0, err
	}
	return uint16(b[0])<<8 | uint16(b[1]), nil
}

func (g *Grinder) writeU16(c modbus.Client, addr uint16, val uint16) error {
	_, err := c.WriteSingleRegister(addr, val)
	return err
}

func (g *Grinder) Restock() error {
	c, close, err := g.client()
	if err != nil {
		return err
	}
	defer close()
	if err := g.writeU16(c, cmdReg, 2); err != nil {
		return err
	}
	for {
		s, err := g.readU16(c, statusReg)
		if err != nil {
			return err
		}
		if s == 0 {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	return nil
}

func (g *Grinder) GrindAuto() error {
	c, close, err := g.client()
	if err != nil {
		return err
	}
	defer close()
	level, err := g.readU16(c, beanLevelReg)
	if err != nil {
		return err
	}
	s, err := g.readU16(c, statusReg)
	if err != nil {
		return err
	}
	if s == 2 || level < 10 {
		if err := g.writeU16(c, cmdReg, 2); err != nil {
			return err
		}
		for {
			s2, err := g.readU16(c, statusReg)
			if err != nil {
				return err
			}
			if s2 == 0 {
				break
			}
			time.Sleep(200 * time.Millisecond)
		}
	}
	if err := g.writeU16(c, cmdReg, 1); err != nil {
		return err
	}
	for i := 0; i < 50; i++ {
		s3, err := g.readU16(c, statusReg)
		if err != nil {
			return err
		}
		if s3 == 0 {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return errors.New("grind_timeout")
}
