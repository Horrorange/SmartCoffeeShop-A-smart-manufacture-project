package gateway

import (
    "encoding/json"
    "io"
    "net/http"
)

type Gateway struct {
    cfg Config
    coffee *CoffeeMachine
    grinder *Grinder
    ice *IceMaker
    robot *DeliveryRobot
}

func New(cfg Config) *Gateway {
    g := &Gateway{
        cfg: cfg,
        coffee: &CoffeeMachine{Host: cfg.CoffeeHost, Port: cfg.CoffeePort},
        grinder: &Grinder{Host: cfg.GrinderHost, Port: cfg.GrinderPort},
        ice: &IceMaker{Host: cfg.IceHost, Rack: cfg.IceRack, Slot: cfg.IceSlot},
        robot: &DeliveryRobot{Host: cfg.MqttHost, Port: cfg.MqttPort},
    }
    if g.grinder.Port == 502 {
        g.grinder.Port = 5021
    }
    return g
}

func (g *Gateway) Grinder() *Grinder { return g.grinder }
func (g *Gateway) Coffee() *CoffeeMachine { return g.coffee }
func (g *Gateway) Ice() *IceMaker { return g.ice }
func (g *Gateway) Robot() *DeliveryRobot { return g.robot }

// HandleHTTP 接收统一JSON指令并分发至具体设备
func (g *Gateway) HandleHTTP(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        w.Write(Result{Ok: false, Message: "method"}.Bytes())
        return
    }
    var cmd UnifiedCommand
    b, err := io.ReadAll(r.Body)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        w.Write(Result{Ok: false, Message: "json"}.Bytes())
        return
    }
    if err := json.Unmarshal(b, &cmd); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        w.Write(Result{Ok: false, Message: "json"}.Bytes())
        return
    }
    res := g.Dispatch(cmd)
    if res.Ok {
        w.WriteHeader(http.StatusOK)
    } else {
        w.WriteHeader(http.StatusBadRequest)
    }
    w.Write(res.Bytes())
}

// Dispatch 转译统一指令为具体设备协议并执行
func (g *Gateway) Dispatch(cmd UnifiedCommand) Result {
    switch cmd.Device {
    case "coffeemachine", "coffee_machine":
        switch cmd.Action {
        case "make":
            if err := g.coffee.Make(cmd.CoffeeType); err != nil {
                return Result{Ok: false, Message: err.Error()}
            }
            return Result{Ok: true, Message: "coffee_done"}
        case "refill_all":
            if err := g.coffee.RefillAll(); err != nil {
                return Result{Ok: false, Message: err.Error()}
            }
            return Result{Ok: true, Message: "refill_ok"}
        default:
            return Result{Ok: false, Message: "unknown_action"}
        }
    case "grinder":
        switch cmd.Action {
        case "grind":
            if err := g.grinder.GrindAuto(); err != nil {
                return Result{Ok: false, Message: err.Error()}
            }
            return Result{Ok: true, Message: "grind_done"}
        case "restock":
            if err := g.grinder.Restock(); err != nil {
                return Result{Ok: false, Message: err.Error()}
            }
            return Result{Ok: true, Message: "restock_ok"}
        default:
            return Result{Ok: false, Message: "unknown_action"}
        }
    case "ice_maker", "icemaker":
        switch cmd.Action {
        case "produce":
            if err := g.ice.ProduceUntil(cmd.IceAmount); err != nil {
                return Result{Ok: false, Message: err.Error()}
            }
            return Result{Ok: true, Message: "ice_ready"}
        case "dispense":
            if err := g.ice.Dispense(cmd.IceAmount); err != nil {
                return Result{Ok: false, Message: err.Error()}
            }
            return Result{Ok: true, Message: "ice_dispensed"}
        default:
            return Result{Ok: false, Message: "unknown_action"}
        }
    case "delivery_robots", "delivery_robot":
        switch cmd.Action {
        case "deliver":
            if err := g.robot.Deliver(cmd.CoffeeType, cmd.NeedIce, cmd.TableNumber); err != nil {
                return Result{Ok: false, Message: err.Error()}
            }
            return Result{Ok: true, Message: "order_sent"}
        default:
            return Result{Ok: false, Message: "unknown_action"}
        }
    default:
        return Result{Ok: false, Message: "unknown_device"}
    }
}