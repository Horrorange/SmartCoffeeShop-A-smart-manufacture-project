package gateway

import (
	"bufio"
	"errors"
	"net"
	"strings"
	"time"
)

type CoffeeMachine struct {
	Host string
	Port int
}

func (c *CoffeeMachine) conn() (net.Conn, error) {
	return net.DialTimeout("tcp", c.addr(), 3*time.Second)
}

func (c *CoffeeMachine) addr() string {
	return c.Host + ":" + itoa(c.Port)
}

func itoa(i int) string {
	return fmtInt(i)
}

func fmtInt(i int) string {
	b := []byte{}
	if i == 0 {
		return "0"
	}
	n := i
	if n < 0 {
		b = append(b, '-')
		n = -n
	}
	s := []byte{}
	for n > 0 {
		d := n % 10
		s = append([]byte{byte('0' + d)}, s...)
		n /= 10
	}
	return string(append(b, s...))
}

func (c *CoffeeMachine) send(cmd string) (string, error) {
	conn, err := c.conn()
	if err != nil {
		return "", err
	}
	defer conn.Close()
	_, err = conn.Write([]byte(cmd + "\n"))
	if err != nil {
		return "", err
	}
	r := bufio.NewReader(conn)
	resp, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(resp), nil
}

func (c *CoffeeMachine) Make(t string) error {
	conn, err := c.conn()
	if err != nil {
		return err
	}
	defer conn.Close()
	if _, err = conn.Write([]byte("MAKE:" + t + "\n")); err != nil {
		return err
	}
	r := bufio.NewReader(conn)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return err
		}
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "ERROR:INSUFFICIENT_INGREDIENT") {
			if err := c.RefillAll(); err != nil {
				return err
			}
			time.Sleep(500 * time.Millisecond)
			if _, err = conn.Write([]byte("MAKE:" + t + "\n")); err != nil {
				return err
			}
			continue
		}
		if strings.HasPrefix(line, "ERROR") {
			return errors.New(line)
		}
		if strings.HasPrefix(line, "DONE:") {
			return nil
		}
	}
}

func (c *CoffeeMachine) RefillAll() error {
	resp, err := c.send("REFILL:ALL")
	if err != nil {
		return err
	}
	if strings.HasPrefix(resp, "ERROR") {
		return errors.New(resp)
	}
	return nil
}
