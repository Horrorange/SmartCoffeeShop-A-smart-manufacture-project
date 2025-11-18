package pipeline

import "time"

// Order 表示一条待处理订单，包含业务字段与流水线状态
// Status: pending→queued→done/error；FinishTime/ErrorMsg 用于记录结果
type Order struct {
    ID int
    CoffeeType string
    BoolIce bool
    TableNum int
    CreateTime time.Time
    Status string
    FinishTime time.Time
    ErrorMsg string
}

// Result 为流水线完成后写回数据库的结果摘要
type Result struct {
    ID int
    BoolIsDone bool
    FinishTime time.Time
    ErrorMsg string
}