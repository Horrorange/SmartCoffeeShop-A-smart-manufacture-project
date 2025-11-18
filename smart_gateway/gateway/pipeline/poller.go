package pipeline

import (
    "context"
    "time"
)

// OrderPoller 周期性从数据库扫描 pending 订单并发布到队列
type OrderPoller struct {
    DB DB
    Q Queue
    Interval time.Duration
    Batch int
}

// Start 启动轮询任务，将命中的订单标记为 queued 后发布
func (p *OrderPoller) Start(ctx context.Context) {
    t := time.NewTicker(p.Interval)
    go func() {
        defer t.Stop()
        for {
            select {
            case <-ctx.Done():
                return
            case <-t.C:
                orders, _ := p.DB.PollPending(p.Batch)
                if len(orders) == 0 { continue }
                ids := make([]int, 0, len(orders))
                for i := range orders { ids = append(ids, orders[i].ID) }
                _ = p.DB.MarkQueued(ids)
                for i := range orders { _ = p.Q.PublishOrder(orders[i]) }
            }
        }
    }()
}