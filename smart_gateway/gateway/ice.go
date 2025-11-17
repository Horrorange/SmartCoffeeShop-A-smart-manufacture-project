package gateway

import (
    "encoding/binary"
    "time"
    gos7 "github.com/robinson/gos7"
)

// IceMaker 通过 S7 协议访问制冰机的 DB1
// DB1 布局（偏移字节）:
// 0: 库存(int)  2: 状态(int)  4: 指令(int)  6: 出冰量(int)
type IceMaker struct {
    Host string
    Rack int
    Slot int
}

func (i *IceMaker) handler() (*gos7.TCPClientHandler, error) {
    h := gos7.NewTCPClientHandler(i.Host, i.Rack, i.Slot)
    h.Timeout = 10 * time.Second
    h.IdleTimeout = 10 * time.Second
    if err := h.Connect(); err != nil {
        return nil, err
    }
    return h, nil
}

func (i *IceMaker) readInt(client gos7.Client, db int, start int) (int, error) {
    buf := make([]byte, 2)
    if err := client.AGReadDB(db, start, 2, buf); err != nil {
        return 0, err
    }
    return int(int16(binary.BigEndian.Uint16(buf))), nil
}

func (i *IceMaker) writeInt(client gos7.Client, db int, start int, v int) error {
    b := make([]byte, 2)
    binary.BigEndian.PutUint16(b, uint16(v))
    return client.AGWriteDB(db, start, 2, b)
}

func (i *IceMaker) ProduceUntil(min int) error {
    h, err := i.handler()
    if err != nil {
        return err
    }
    defer h.Close()
    c := gos7.NewClient(h)
    inv, err := i.readInt(c, 1, 0)
    if err != nil {
        return err
    }
    if inv >= min {
        return nil
    }
    if err := i.writeInt(c, 1, 4, 1); err != nil {
        return err
    }
    for {
        inv, err = i.readInt(c, 1, 0)
        if err != nil {
            return err
        }
        if inv >= min {
            break
        }
        time.Sleep(500 * time.Millisecond)
    }
    return nil
}

func (i *IceMaker) Dispense(amount int) error {
    h, err := i.handler()
    if err != nil {
        return err
    }
    defer h.Close()
    c := gos7.NewClient(h)
    if err := i.writeInt(c, 1, 6, amount); err != nil {
        return err
    }
    if err := i.writeInt(c, 1, 4, 3); err != nil {
        return err
    }
    for {
        st, err := i.readInt(c, 1, 2)
        if err != nil {
            return err
        }
        if st == 0 {
            break
        }
        time.Sleep(200 * time.Millisecond)
    }
    return nil
}