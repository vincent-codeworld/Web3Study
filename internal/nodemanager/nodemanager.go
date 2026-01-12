package nodemanager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

type NodesManager struct {
	nodes        []*Node
	currentIndex int
	//mu                  sync.RWMutex
	healthCheckInterval time.Duration
	maxRetries          int
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

func NewNodesManager(configs []NodeConfig) (*NodesManager, error) {
	if len(configs) == 0 {
		return nil, fmt.Errorf("no config provided")
	}
	ctx, cancelFunc := context.WithCancel(context.Background())
	nm := &NodesManager{
		nodes:               make([]*Node, 0, len(configs)),
		ctx:                 ctx,
		cancel:              cancelFunc,
		maxRetries:          3,
		healthCheckInterval: 10 * time.Second,
	}
	index := 0
	for _, config := range configs {
		if node, err := nm.createNode(config); err != nil {
			//todo print out error logs
			continue
		} else {
			nm.nodes[index] = node
			index++
		}
	}
	nm.nodes = nm.nodes[:index]
	go nm.startNodesStatusCheck()
	return nm, nil
}

func (nm *NodesManager) createNode(config NodeConfig) (*Node, error) {
	client, err := ethclient.Dial(config.Url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum client: %v", err)
	}
	if config.TimeOut == 0 {
		config.TimeOut = 10 * time.Second
	}

	return &Node{
		Client: client,
		Config: config,
		Status: NodeStatus{
			IsHealthy:     true,
			LastCheckTime: time.Now(),
		},
	}, nil
}

func (nm *NodesManager) GetHealthNode() (*Node, error) {
	/*	nm.mu.RLock()
		defer nm.mu.RUnlock()*/
	var bestNode *Node
	for _, node := range nm.nodes {
		node.mu.RLock()
		isHealthy := node.Status.IsHealthy
		priority := node.Config.Priority
		node.mu.RUnlock()
		if !isHealthy {
			continue
		}
		if bestNode == nil {
			bestNode = node
			continue
		}
		if priority < bestNode.Config.Priority {
			bestNode = node
		}
	}

	if bestNode == nil {
		return nil, fmt.Errorf("could not find a healthy node")
	}
	return bestNode, nil
}

func (nm *NodesManager) checkNodeStatus(node *Node) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), node.Config.TimeOut)
	defer cancelFunc()
	startTime := time.Now()
	_, err := node.Client.BlockNumber(ctx)
	responseTime := time.Since(startTime)
	node.mu.Lock()
	defer node.mu.Unlock()
	if err != nil {
		node.Status.ErrorCount++
		node.Status.LastError = err
		if node.Status.ErrorCount >= 3 {
			node.Status.IsHealthy = false
			//todo print out logs
		}
		return
	}
	//  status checked is success
	node.Status.LastError = nil
	node.Status.ResponseTime = responseTime
	node.Status.SuccessCount++
	node.Status.ErrorCount = 0
	// print out some logs about status for the node
}

func (nm *NodesManager) checkAllNodes() {
	var wg sync.WaitGroup

	for _, node := range nm.nodes {
		wg.Add(1)
		go func(n *Node) {
			defer wg.Done()
			nm.checkNodeStatus(n)
		}(node)
	}
	wg.Wait()
}

func (nm *NodesManager) startNodesStatusCheck() {
	timer := time.NewTimer(nm.healthCheckInterval)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			{
				nm.checkAllNodes()
				timer.Reset(nm.healthCheckInterval)
			}
		case <-nm.ctx.Done():
			return
		}
	}
}

func (nm *NodesManager) Close() {
	nm.cancel()
	for _, node := range nm.nodes {
		node.mu.Lock()
		if node.Client != nil {
			node.Client.Close()
		}
		node.mu.Unlock()
	}
}
func (nm *NodesManager) ExecuteWithRetry(fn func(client *ethclient.Client) error) error {
	var lastErr error
	for attempts := 0; attempts < nm.maxRetries; attempts++ {
		node, err := nm.GetHealthNode()
		if err != nil {
			lastErr = err
			time.Sleep(time.Second * time.Duration(attempts+1))
			//todo print out logs
			continue
		}
		err = fn(node.Client)
		if err != nil {
			lastErr = err
			node.mu.Lock()
			node.Status.LastError = err
			node.Status.ErrorCount++
			if node.Status.ErrorCount >= 3 {
				node.Status.IsHealthy = false
				//todo print out logs
			}
			node.mu.Unlock()
			lastErr = err
			time.Sleep(time.Second * time.Duration(attempts+1))
			continue
		}
		node.mu.Lock()
		node.Status.SuccessCount++
		node.mu.Unlock()
		return nil
	}

	return fmt.Errorf("max retries exceeded: %v", lastErr)
}

// todo
func (nm *NodesManager) GetNodeStatss() []map[string]interface{} {
	return nil
}
