package pipeline

import (
    "sync"
    "time"
)

// DB 抽象真实数据库接口，便于替换为 PostgreSQL/MySQL 实现
type DB interface {
    // PollPending 拉取一定数量的 pending 订单
    PollPending(limit int) ([]Order, error)
    // MarkQueued 将拉取的订单原子标记为 queued，防止重复处理
    MarkQueued(ids []int) error
    // UpdateResult 按 id 写回完成/失败结果
    UpdateResult(res Result) error
    // Seed/All 为演示用的内存数据注入与读取
    Seed(orders []Order)
    All() []Order
}

// InMemoryDB 为演示用的内存数据库，实现 DB 接口
type InMemoryDB struct {
    mu sync.Mutex
    orders []Order
}

func NewInMemoryDB() *InMemoryDB { return &InMemoryDB{} }

// Seed 注入初始订单，设置默认时间与状态
func (d *InMemoryDB) Seed(orders []Order) {
    d.mu.Lock(); defer d.mu.Unlock()
    for i := range orders {
        if orders[i].CreateTime.IsZero() { orders[i].CreateTime = time.Now() }
        if orders[i].Status == "" { orders[i].Status = "pending" }
        d.orders = append(d.orders, orders[i])
    }
}

// PollPending 返回最多 limit 条 pending 订单
func (d *InMemoryDB) PollPending(limit int) ([]Order, error) {
    d.mu.Lock(); defer d.mu.Unlock()
    res := []Order{}
    for i := range d.orders {
        if d.orders[i].Status == "pending" {
            res = append(res, d.orders[i])
            if len(res) >= limit { break }
        }
    }
    return res, nil
}

// MarkQueued 将指定 id 的订单标记为 queued
func (d *InMemoryDB) MarkQueued(ids []int) error {
    d.mu.Lock(); defer d.mu.Unlock()
    idset := map[int]struct{}{}
    for _, id := range ids { idset[id] = struct{}{} }
    for i := range d.orders {
        if _, ok := idset[d.orders[i].ID]; ok && d.orders[i].Status == "pending" {
            d.orders[i].Status = "queued"
        }
    }
    return nil
}

// UpdateResult 写回完成/失败状态与时间、错误信息
func (d *InMemoryDB) UpdateResult(res Result) error {
    d.mu.Lock(); defer d.mu.Unlock()
    for i := range d.orders {
        if d.orders[i].ID == res.ID {
            if res.BoolIsDone { d.orders[i].Status = "done" } else { d.orders[i].Status = "error" }
            d.orders[i].FinishTime = res.FinishTime
            d.orders[i].ErrorMsg = res.ErrorMsg
            return nil
        }
    }
    return nil
}

// All 返回所有订单（演示输出使用）
func (d *InMemoryDB) All() []Order {
    d.mu.Lock(); defer d.mu.Unlock()
    out := make([]Order, len(d.orders))
    copy(out, d.orders)
    return out
}