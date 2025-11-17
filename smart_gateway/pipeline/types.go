package pipeline

import "time"

type Order struct {
    ID int
    CoffeeType string
    BoolIce bool
    TableNum int
    CreateTime time.Time
    Status string
}

type Result struct {
    ID int
    BoolIsDone bool
    FinishTime time.Time
    ErrorMsg string
}