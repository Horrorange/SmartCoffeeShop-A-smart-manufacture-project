package pipeline

import "context"

type ResultSyncer struct {
    DB DB
    Q Queue
}

func (s *ResultSyncer) Start(ctx context.Context) {
    results := s.Q.ConsumeResults(ctx)
    go func() {
        for r := range results {
            _ = s.DB.UpdateResult(r)
        }
    }()
}