package pipeline

import (
    "context"
    "time"
)

type OrderPoller struct {
    DB DB
    Q Queue
    Interval time.Duration
    Batch int
}

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
                if len(orders) == 0 {
                    continue
                }
                ids := make([]int, 0, len(orders))
                for i := range orders { ids = append(ids, orders[i].ID) }
                _ = p.DB.MarkQueued(ids)
                for i := range orders { _ = p.Q.PublishOrder(orders[i]) }
            }
        }
    }()
}