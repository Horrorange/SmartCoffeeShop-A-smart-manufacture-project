package main

import (
    "fmt"
    "log"
    "net/http"
    "os"
    "smart_gateway/gateway"
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
    mux := http.NewServeMux()
    mux.HandleFunc("/cmd", g.HandleHTTP)
    srv := &http.Server{Addr: ":9090", Handler: mux}
    log.Fatal(srv.ListenAndServe())
}

func env(k, d string) string {
    v := os.Getenv(k)
    if v == "" {
        return d
    }
    return v
}

func envInt(k string, d int) int {
    v := os.Getenv(k)
    if v == "" {
        return d
    }
    var x int
    _, err := fmt.Sscanf(v, "%d", &x)
    if err != nil {
        return d
    }
    return x
}