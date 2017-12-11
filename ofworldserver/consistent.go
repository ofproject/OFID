package ofworldserver

import (
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

const (
	DEFAULT_REPLICAS = 1000
)

type HashRing []uint32

func (h HashRing) Len() int {
	return len(h)
}

func (h HashRing) Less(i, j int) bool {
	return h[i] < h[j]
}

func (h HashRing) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

type Consistent struct {
	Nodes     map[uint32]*Node
	numReps   int
	Resources map[string]bool
	ring      HashRing
	sync.RWMutex
}

var (
	globConsistent *Consistent = nil
)

func NewConsistent() *Consistent {
	if globConsistent == nil {
		globConsistent = &Consistent{
			Nodes:     make(map[uint32]*Node),
			numReps:   DEFAULT_REPLICAS,
			Resources: make(map[string]bool),
			ring:      HashRing{},
		}
	}

	return globConsistent
}

func (c *Consistent) joinStr(i int, node *Node) string {
	return node.Id + "*" + strconv.Itoa(node.Weight) + "-" + strconv.Itoa(i)
}

func (c *Consistent) sortHashRing() {
	c.ring = HashRing{}
	for k := range c.Nodes {
		c.ring = append(c.ring, k)
	}

	sort.Sort(c.ring)
}

func (c *Consistent) hashStr(key string) uint32 {
	return crc32.ChecksumIEEE([]byte(key))
}

func (c *Consistent) Add(node *Node) bool {
	c.Lock()
	defer c.Unlock()

	if _, ok := c.Resources[node.Id]; ok {
		Logger.Warning("Node already exist id: ", node.Id)
		return false
	}

	count := c.numReps * node.Weight
	for i := 0; i < count; i++ {
		str := c.joinStr(i, node)
		c.Nodes[c.hashStr(str)] = node
	}
	c.Resources[node.Id] = true
	c.sortHashRing()
	return true
}

func (c *Consistent) search(hash uint32) int {
	i := sort.Search(len(c.ring), func(i int) bool { return c.ring[i] >= hash })
	if i < len(c.ring) {
		if i == len(c.ring)-1 {
			return 0
		} else {
			return i
		}
	} else {
		return len(c.ring) - 1
	}
}

func (c *Consistent) Get(key string) (*Node, uint32) {
	if c.ring.Len() <= 0 {
		return nil, 0
	}

	c.RLock()
	defer c.RUnlock()

	hash := c.hashStr(key)
	i := c.search(hash)
	return c.Nodes[c.ring[i]], c.ring[i]
}

func (c *Consistent) GetFromKey(key uint32) *Node {
	c.RLock()
	defer c.RUnlock()

	i := c.search(key)
	if c.ring[i] != key {
		return nil
	}

	return c.Nodes[key]
}

func (c *Consistent) Remove(node *Node) {
	c.Lock()
	defer c.Unlock()

	if _, ok := c.Resources[node.Id]; !ok {
		return
	}

	delete(c.Resources, node.Id)

	count := c.numReps * node.Weight
	for i := 0; i < count; i++ {
		str := c.joinStr(i, node)
		delete(c.Nodes, c.hashStr(str))
	}

	c.sortHashRing()
}
