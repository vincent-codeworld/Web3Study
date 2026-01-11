package nodemanager

import (
	"context"
	"github.com/ethereum/go-ethereum/ethclient"
	"sync"
	"time"
)

type NodesManager struct {
	nodes               []*Node
	currentIndex        int
	mu                  sync.RWMutex
	healthCheckInterval time.Duration
	ctx                 context.Context
	cancel              context.CancelFunc
}

type Node struct {
	Client *ethclient.Client
	Config NodeConfig
	Status NodeStatus
	mu     sync.RWMutex
}

type NodeConfig struct {
	Name     string
	Url      string
	Priority int
	Weight   int
	TimeOut  time.Duration
}
type NodeStatus struct {
	IsHealthy     bool
	LastCheckTime time.Time
	ResponseTime  time.Duration
	ErrorCount    int
	SuccessCount  int
	LastError     error
}
