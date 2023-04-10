package hinge

import (
	"sync"
)

func NewHinge() *Hinge {
	return &Hinge{
		clusters: make(map[string]*Cluster),
	}
}

type Hinge struct {
	clusters map[string]*Cluster
	lock     sync.Mutex
}

func (h *Hinge) AddCluster(id string, cls *Cluster) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.clusters[id] = cls
}

func (h *Hinge) GetCluster(id string) *Cluster {
	h.lock.Lock()
	defer h.lock.Unlock()
	return h.clusters[id]
}

func (h *Hinge) DelCluster(id string) {
	h.lock.Lock()
	defer h.lock.Unlock()
	delete(h.clusters, id)
}
