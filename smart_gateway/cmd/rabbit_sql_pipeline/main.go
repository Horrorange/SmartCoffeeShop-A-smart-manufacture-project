package main

import (
    "context"
    "fmt"
    "os"
    "smart_gateway/gateway"
    gp "smart_gateway/gateway/pipeline"
    "time"
)

func main() {
    cfg := gateway.Config{
        CoffeeHost: env("COFFEE_HOST", "localhost"),
        CoffeePort: envInt("COFFEE_PORT", 8888),
        GrinderHost: env("GRINDER_HOST", "localhost"),
        GrinderPort: envInt("GRINDER_PORT", 5021),
        IceHost: env("ICE_HOST", "localhost"),
        IceRack: envInt("ICE_RACK", 0),
        IceSlot: envInt("ICE_SLOT", 2),
        MqttHost: env("MQTT_HOST", "localhost"),
        MqttPort: envInt("MQTT_PORT", 1883),
    }
    g := gateway.New(cfg)

    db, err := gp.NewSQLDB()
    if err != nil { panic(err) }
    q, err := gp.NewAMQPQueue()
    if err != nil { panic(err) }

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    iceMin := envInt("ICE_MIN_STOCK", 200)
    iceDispense := envInt("ICE_DISPENSE_AMOUNT", 100)

    poller := &gp.OrderPoller{DB: db, Q: q, Interval: 500 * time.Millisecond, Batch: 50}
    core := gp.NewCore(q, g.Grinder(), g.Coffee(), g.Ice(), g.Robot(), iceMin, iceDispense)
    syncer := &gp.ResultSyncer{DB: db, Q: q}

    poller.Start(ctx)
    core.Start(ctx)
    syncer.Start(ctx)

    for {
        time.Sleep(2 * time.Second)
        fmt.Println("-- progress snapshot --")
        for _, o := range db.All() {
            if o.Status == "error" {
                fmt.Println(o.ID, o.CoffeeType, o.BoolIce, o.TableNum, o.Status, o.ErrorMsg)
            } else {
                fmt.Println(o.ID, o.CoffeeType, o.BoolIce, o.TableNum, o.Status)
            }
        }
    }
}

func env(k, d string) string { v := os.Getenv(k); if v == "" { return d }; return v }
func envInt(k string, d int) int { v := os.Getenv(k); if v == "" { return d }; var x int; _, err := fmt.Sscanf(v, "%d", &x); if err != nil { return d }; return x }