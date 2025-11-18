package pipeline

import (
    "context"
    "fmt"
    "smart_gateway/gateway"
    "time"
)

// Core 为流水线核心，绑定 gateway 的真实设备驱动并通过信号量控制互斥资源
type Core struct {
    Q Queue
    Grinder *gateway.Grinder
    Coffee *gateway.CoffeeMachine
    Ice *gateway.IceMaker
    Robot *gateway.DeliveryRobot
    grinder chan struct{}
    brewer chan struct{}
    icemaker chan struct{}
    robot chan struct{}
    IceMin int
    IceDispense int
}

// NewCore 初始化四个资源的信号量（容量=1），确保同类设备串行占用
func NewCore(q Queue, g *gateway.Grinder, c *gateway.CoffeeMachine, i *gateway.IceMaker, r *gateway.DeliveryRobot, iceMin, iceDispense int) *Core {
    core := &Core{Q: q, Grinder: g, Coffee: c, Ice: i, Robot: r, IceMin: iceMin, IceDispense: iceDispense}
    core.grinder = make(chan struct{}, 1)
    core.brewer = make(chan struct{}, 1)
    core.icemaker = make(chan struct{}, 1)
    core.robot = make(chan struct{}, 1)
    core.grinder <- struct{}{}
    core.brewer <- struct{}{}
    core.icemaker <- struct{}{}
    core.robot <- struct{}{}
    return core
}

// Start 消费订单队列，为每条订单启动一个 goroutine 进行流水线处理
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

// process 执行四步工序，任一步失败都会发布 error 结果并停止该订单
func (c *Core) process(ctx context.Context, o Order) {
    fmt.Println("order_start", o.ID, o.CoffeeType, o.BoolIce, o.TableNum)
    if err := c.stepGrind(); err != nil { fmt.Println("grind_error", o.ID, err.Error()); _ = c.Q.PublishResult(Result{ID: o.ID, BoolIsDone: false, FinishTime: time.Now(), ErrorMsg: err.Error()}); return }
    fmt.Println("grind_done", o.ID)
    if err := c.stepBrew(o.CoffeeType); err != nil { fmt.Println("brew_error", o.ID, err.Error()); _ = c.Q.PublishResult(Result{ID: o.ID, BoolIsDone: false, FinishTime: time.Now(), ErrorMsg: err.Error()}); return }
    fmt.Println("brew_done", o.ID, o.CoffeeType)
    if o.BoolIce {
        if err := c.stepIce(); err != nil { fmt.Println("ice_error", o.ID, err.Error()); _ = c.Q.PublishResult(Result{ID: o.ID, BoolIsDone: false, FinishTime: time.Now(), ErrorMsg: err.Error()}); return }
        fmt.Println("ice_done", o.ID, c.IceDispense)
    }
    if err := c.stepDeliver(o); err != nil {
        if err.Error() == "deliver_timeout" {
            fmt.Println("deliver_timeout_done", o.ID)
            _ = c.Q.PublishResult(Result{ID: o.ID, BoolIsDone: true, FinishTime: time.Now()})
            return
        }
        fmt.Println("deliver_error", o.ID, err.Error())
        _ = c.Q.PublishResult(Result{ID: o.ID, BoolIsDone: false, FinishTime: time.Now(), ErrorMsg: err.Error()})
        return
    }
    fmt.Println("deliver_done", o.ID, o.TableNum)
    _ = c.Q.PublishResult(Result{ID: o.ID, BoolIsDone: true, FinishTime: time.Now()})
    fmt.Println("order_done", o.ID)
}

// stepGrind 占用磨豆机并进行自动磨豆
func (c *Core) stepGrind() error {
    c.acquire(c.grinder); defer c.release(c.grinder)
    return c.Grinder.GrindAuto()
}

// stepBrew 占用咖啡机进行制作
func (c *Core) stepBrew(coffeeType string) error {
    c.acquire(c.brewer); defer c.release(c.brewer)
    return c.Coffee.Make(coffeeType)
}

// stepIce 占用制冰机，先补足库存再按指定量出冰
func (c *Core) stepIce() error {
    c.acquire(c.icemaker); defer c.release(c.icemaker)
    if err := c.Ice.ProduceUntil(c.IceMin); err != nil { return err }
    return c.Ice.Dispense(c.IceDispense)
}

// stepDeliver 占用送餐机器人并发送配送任务（3s 内未回执视为超时）
func (c *Core) stepDeliver(o Order) error {
    c.acquire(c.robot); defer c.release(c.robot)
    return c.Robot.Deliver(o.CoffeeType, o.BoolIce, o.TableNum)
}