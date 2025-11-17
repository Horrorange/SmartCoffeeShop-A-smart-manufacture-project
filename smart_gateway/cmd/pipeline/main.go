package main

import (
    "context"
    "fmt"
    "smart_gateway/pipeline"
    "time"
)

func main() {
    db := pipeline.NewInMemoryDB()
    db.Seed([]pipeline.Order{
        {ID: 1, CoffeeType: "LATTE", BoolIce: true, TableNum: 1},
        {ID: 2, CoffeeType: "AMERICANO", BoolIce: false, TableNum: 2},
        {ID: 3, CoffeeType: "ESPRESSO", BoolIce: false, TableNum: 3},
        {ID: 4, CoffeeType: "MOCHA", BoolIce: true, TableNum: 4},
        {ID: 5, CoffeeType: "CAPPUCCINO", BoolIce: true, TableNum: 5},
    })
    q := pipeline.NewInMemoryQueue(100)
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    poller := &pipeline.OrderPoller{DB: db, Q: q, Interval: 500 * time.Millisecond, Batch: 10}
    core := pipeline.NewCore(q)
    syncer := &pipeline.ResultSyncer{DB: db, Q: q}
    poller.Start(ctx)
    core.Start(ctx)
    syncer.Start(ctx)
    time.Sleep(40 * time.Second)
    for _, o := range db.All() {
        fmt.Println(o.ID, o.CoffeeType, o.BoolIce, o.TableNum, o.Status)
    }
}