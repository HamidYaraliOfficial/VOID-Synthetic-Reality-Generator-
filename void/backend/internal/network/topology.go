// Package network implements VOID's Network Topology Simulator: servers,
// routers, services, clients and databases modeled as a dependency graph
// with latency, bandwidth, packet loss and failure/recovery propagation.
package network

import "sync"

type NodeKind string

const (
	NodeServer   NodeKind = "server"
	NodeRouter   NodeKind = "router"
	NodeService  NodeKind = "service"
	NodeClient   NodeKind = "client"
	NodeDatabase NodeKind = "database"
)

type Node struct {
	ID          string
	Kind        NodeKind
	LatencyMs   float64
	BandwidthMbps float64
	PacketLossPct float64
	Healthy     bool
	DependsOn   []string // node IDs this node requires to function
}

// Topology is a directed dependency graph of Nodes.
type Topology struct {
	mu    sync.RWMutex
	Nodes map[string]*Node
}

func NewTopology() *Topology { return &Topology{Nodes: map[string]*Node{}} }

func (t *Topology) AddNode(n *Node) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if n.Healthy == false && n.PacketLossPct == 0 {
		n.Healthy = true
	}
	t.Nodes[n.ID] = n
}

func (t *Topology) Fail(id string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if n, ok := t.Nodes[id]; ok {
		n.Healthy = false
	}
}

func (t *Topology) Recover(id string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if n, ok := t.Nodes[id]; ok {
		n.Healthy = true
	}
}

// IsAvailable reports whether a node - and transitively, everything it
// depends on - is currently healthy (a naive but effective cascading
// dependency-failure model, cycle-safe via a visited set).
func (t *Topology) IsAvailable(id string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	visited := map[string]bool{}
	return t.isAvailableLocked(id, visited)
}

func (t *Topology) isAvailableLocked(id string, visited map[string]bool) bool {
	if visited[id] {
		return true // break cycles optimistically
	}
	visited[id] = true
	n, ok := t.Nodes[id]
	if !ok || !n.Healthy {
		return false
	}
	for _, dep := range n.DependsOn {
		if !t.isAvailableLocked(dep, visited) {
			return false
		}
	}
	return true
}

// CascadeFailures returns every node whose availability is currently false,
// directly or transitively, useful for the Chaos impact report.
func (t *Topology) CascadeFailures() []string {
	t.mu.RLock()
	ids := make([]string, 0, len(t.Nodes))
	for id := range t.Nodes {
		ids = append(ids, id)
	}
	t.mu.RUnlock()

	var down []string
	for _, id := range ids {
		if !t.IsAvailable(id) {
			down = append(down, id)
		}
	}
	return down
}
