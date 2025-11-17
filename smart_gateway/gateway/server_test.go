package gateway

import "testing"

func TestGrinderPortFallback(t *testing.T) {
    cfg := Config{GrinderHost: "localhost", GrinderPort: 502}
    g := New(cfg)
    if g.grinder.Port != 5021 {
        t.Fatalf("expected grinder port 5021, got %d", g.grinder.Port)
    }
}