package pipeline

import "context"

// Queue 抽象队列接口，便于替换为 RabbitMQ/AMQP 实现
type Queue interface {
    PublishOrder(o Order) error
    ConsumeOrders(ctx context.Context) <-chan Order
    PublishResult(r Result) error
    ConsumeResults(ctx context.Context) <-chan Result
}

// InMemoryQueue 为演示用的内存队列，实现 Queue 接口
type InMemoryQueue struct {
    orders chan Order
    results chan Result
}

func NewInMemoryQueue(buffer int) *InMemoryQueue {
    return &InMemoryQueue{orders: make(chan Order, buffer), results: make(chan Result, buffer)}
}

func (q *InMemoryQueue) PublishOrder(o Order) error { q.orders <- o; return nil }
func (q *InMemoryQueue) PublishResult(r Result) error { q.results <- r; return nil }

// ConsumeOrders 返回一个只读通道，持续输出订单消息
func (q *InMemoryQueue) ConsumeOrders(ctx context.Context) <-chan Order {
    out := make(chan Order)
    go func() {
        defer close(out)
        for {
            select {
            case <-ctx.Done():
                return
            case o := <-q.orders:
                out <- o
            }
        }
    }()
    return out
}

// ConsumeResults 返回一个只读通道，持续输出处理结果
func (q *InMemoryQueue) ConsumeResults(ctx context.Context) <-chan Result {
    out := make(chan Result)
    go func() {
        defer close(out)
        for {
            select {
            case <-ctx.Done():
                return
            case r := <-q.results:
                out <- r
            }
        }
    }()
    return out
}