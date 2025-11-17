package pipeline

import (
    "sync"
    "time"
)

type DB interface {
    PollPending(limit int) ([]Order, error)
    MarkQueued(ids []int) error
    UpdateResult(res Result) error
    Seed(orders []Order)
    All() []Order
}

type InMemoryDB struct {
    mu sync.Mutex
    orders []Order
}

func NewInMemoryDB() *InMemoryDB {
    return &InMemoryDB{}
}

func (d *InMemoryDB) Seed(orders []Order) {
    d.mu.Lock()
    defer d.mu.Unlock()
    for i := range orders {
        if orders[i].CreateTime.IsZero() {
            orders[i].CreateTime = time.Now()
        }
        if orders[i].Status == "" {
            orders[i].Status = "pending"
        }
        d.orders = append(d.orders, orders[i])
    }
}

func (d *InMemoryDB) PollPending(limit int) ([]Order, error) {
    d.mu.Lock()
    defer d.mu.Unlock()
    res := []Order{}
    for i := range d.orders {
        if d.orders[i].Status == "pending" {
            res = append(res, d.orders[i])
            if len(res) >= limit {
                break
            }
        }
    }
    return res, nil
}

func (d *InMemoryDB) MarkQueued(ids []int) error {
    d.mu.Lock()
    defer d.mu.Unlock()
    idset := map[int]struct{}{}
    for _, id := range ids {
        idset[id] = struct{}{}
    }
    for i := range d.orders {
        if _, ok := idset[d.orders[i].ID]; ok && d.orders[i].Status == "pending" {
            d.orders[i].Status = "queued"
        }
    }
    return nil
}

func (d *InMemoryDB) UpdateResult(res Result) error {
    d.mu.Lock()
    defer d.mu.Unlock()
    for i := range d.orders {
        if d.orders[i].ID == res.ID {
            if res.BoolIsDone {
                d.orders[i].Status = "done"
            } else {
                d.orders[i].Status = "error"
            }
            d.orders[i].FinishTime = res.FinishTime
            d.orders[i].ErrorMsg = res.ErrorMsg
            return nil
        }
    }
    return nil
}

func (d *InMemoryDB) All() []Order {
    d.mu.Lock()
    defer d.mu.Unlock()
    out := make([]Order, len(d.orders))
    copy(out, d.orders)
    return out
}