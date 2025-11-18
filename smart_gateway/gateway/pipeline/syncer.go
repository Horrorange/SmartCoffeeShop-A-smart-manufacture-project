package pipeline

import "context"

// ResultSyncer 监听完成队列并写回数据库
type ResultSyncer struct {
    DB DB
    Q Queue
}

// Start 消费结果并更新对应订单的状态/时间/错误信息
func (s *ResultSyncer) Start(ctx context.Context) {
    results := s.Q.ConsumeResults(ctx)
    go func() {
        for r := range results {
            _ = s.DB.UpdateResult(r)
        }
    }()
}