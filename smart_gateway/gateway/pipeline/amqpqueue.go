package pipeline

import (
    "context"
    "encoding/json"
    amqp "github.com/rabbitmq/amqp091-go"
)

type AMQPQueue struct {
    conn *amqp.Connection
    ch *amqp.Channel
    ordersQ string
    resultsQ string
}

func NewAMQPQueue() (*AMQPQueue, error) {
    url := getenv("AMQP_URL", "amqp://guest:guest@localhost:5672/")
    qOrders := getenv("AMQP_QUEUE_ORDERS", "queue_orders")
    qCompleted := getenv("AMQP_QUEUE_COMPLETED", "queue_completed")
    conn, err := amqp.Dial(url)
    if err != nil { return nil, err }
    ch, err := conn.Channel()
    if err != nil { conn.Close(); return nil, err }
    if _, err = ch.QueueDeclare(qOrders, true, false, false, false, nil); err != nil { ch.Close(); conn.Close(); return nil, err }
    if _, err = ch.QueueDeclare(qCompleted, true, false, false, false, nil); err != nil { ch.Close(); conn.Close(); return nil, err }
    return &AMQPQueue{conn: conn, ch: ch, ordersQ: qOrders, resultsQ: qCompleted}, nil
}

func (q *AMQPQueue) PublishOrder(o Order) error {
    b, _ := json.Marshal(o)
    return q.ch.PublishWithContext(context.Background(), "", q.ordersQ, false, false, amqp.Publishing{ContentType: "application/json", DeliveryMode: amqp.Persistent, Body: b})
}

func (q *AMQPQueue) PublishResult(r Result) error {
    b, _ := json.Marshal(r)
    return q.ch.PublishWithContext(context.Background(), "", q.resultsQ, false, false, amqp.Publishing{ContentType: "application/json", DeliveryMode: amqp.Persistent, Body: b})
}

func (q *AMQPQueue) ConsumeOrders(ctx context.Context) <-chan Order {
    out := make(chan Order)
    go func() {
        defer close(out)
        msgs, err := q.ch.Consume(q.ordersQ, "", true, false, false, false, nil)
        if err != nil { return }
        for {
            select {
            case <-ctx.Done():
                return
            case m, ok := <-msgs:
                if !ok { return }
                var o Order
                if err := json.Unmarshal(m.Body, &o); err == nil { out <- o }
            }
        }
    }()
    return out
}

func (q *AMQPQueue) ConsumeResults(ctx context.Context) <-chan Result {
    out := make(chan Result)
    go func() {
        defer close(out)
        msgs, err := q.ch.Consume(q.resultsQ, "", true, false, false, false, nil)
        if err != nil { return }
        for {
            select {
            case <-ctx.Done():
                return
            case m, ok := <-msgs:
                if !ok { return }
                var r Result
                if err := json.Unmarshal(m.Body, &r); err == nil { out <- r }
            }
        }
    }()
    return out
}