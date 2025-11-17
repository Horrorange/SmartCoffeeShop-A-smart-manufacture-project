package pipeline

import (
    "context"
)

type Queue interface {
    PublishOrder(o Order) error
    ConsumeOrders(ctx context.Context) <-chan Order
    PublishResult(r Result) error
    ConsumeResults(ctx context.Context) <-chan Result
}

type InMemoryQueue struct {
    orders chan Order
    results chan Result
}

func NewInMemoryQueue(buffer int) *InMemoryQueue {
    return &InMemoryQueue{orders: make(chan Order, buffer), results: make(chan Result, buffer)}
}

func (q *InMemoryQueue) PublishOrder(o Order) error {
    q.orders <- o
    return nil
}

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

func (q *InMemoryQueue) PublishResult(r Result) error {
    q.results <- r
    return nil
}

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