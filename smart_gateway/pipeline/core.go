package pipeline

import (
    "context"
    "time"
)

type Core struct {
    Q Queue
    grinder chan struct{}
    brewer chan struct{}
    icemaker chan struct{}
    robot chan struct{}
}

func NewCore(q Queue) *Core {
    c := &Core{Q: q}
    c.grinder = make(chan struct{}, 1)
    c.brewer = make(chan struct{}, 1)
    c.icemaker = make(chan struct{}, 1)
    c.robot = make(chan struct{}, 1)
    c.grinder <- struct{}{}
    c.brewer <- struct{}{}
    c.icemaker <- struct{}{}
    c.robot <- struct{}{}
    return c
}

func (c *Core) Start(ctx context.Context) {
    orders := c.Q.ConsumeOrders(ctx)
    go func() {
        for o := range orders {
            o := o
            go c.process(ctx, o)
        }
    }()
}

func (c *Core) acquire(ch chan struct{}) { <-ch }
func (c *Core) release(ch chan struct{}) { ch <- struct{}{} }

func (c *Core) process(ctx context.Context, o Order) {
    c.acquire(c.grinder)
    time.Sleep(2 * time.Second)
    c.release(c.grinder)
    c.acquire(c.brewer)
    time.Sleep(5 * time.Second)
    c.release(c.brewer)
    if o.BoolIce {
        c.acquire(c.icemaker)
        time.Sleep(1 * time.Second)
        c.release(c.icemaker)
    }
    c.acquire(c.robot)
    time.Sleep(3 * time.Second)
    c.release(c.robot)
    _ = c.Q.PublishResult(Result{ID: o.ID, BoolIsDone: true, FinishTime: time.Now()})
}