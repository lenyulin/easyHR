package snowflake

import (
	"fmt"
	"sync"
	"time"
)

// Snowflake 结构体
type Node struct {
	mu        sync.Mutex
	timestamp int64
	nodeID    int64
	step      int64
}

const (
	nodeBits  uint8 = 10
	stepBits  uint8 = 12
	nodeMax   int64 = -1 ^ (-1 << nodeBits)
	stepMax   int64 = -1 ^ (-1 << stepBits)
	timeShift uint8 = nodeBits + stepBits
	nodeShift uint8 = stepBits
	epoch     int64 = 1577836800000 // 2020-01-01 00:00:00 UTC
)

// NewNode 创建一个新的 Snowflake 节点
func NewNode(nodeID int64) (*Node, error) {
	if nodeID < 0 || nodeID > nodeMax {
		return nil, fmt.Errorf("node ID must be between 0 and %d", nodeMax)
	}
	return &Node{
		timestamp: 0,
		nodeID:    nodeID,
		step:      0,
	}, nil
}

// Generate 生成一个新的 ID
func (n *Node) Generate() int64 {
	n.mu.Lock()
	defer n.mu.Unlock()

	now := time.Now().UnixMilli()

	if n.timestamp == now {
		n.step = (n.step + 1) & stepMax
		if n.step == 0 {
			for now <= n.timestamp {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		n.step = 0
	}

	n.timestamp = now

	return (now-epoch)<<timeShift | (n.nodeID << nodeShift) | (n.step)
}
