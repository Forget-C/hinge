package store

import (
	"fmt"

	"github.com/Forget-C/hinge/index"
	"k8s.io/client-go/tools/cache"
)

type Store interface {
	AddIndexer(indexers cache.Indexers) error
	Run()

	Down()
	IndexFilter(ss string, index index.IndexName) ([]interface{}, error)

	Get(name string, namespace string) (interface{}, bool, error)
	List() []interface{}
	Informer() cache.SharedIndexInformer
}

func NewStore(informer cache.SharedIndexInformer) Store {
	return &store{stop: make(chan struct{}), informer: informer}
}

type store struct {
	informer cache.SharedIndexInformer
	stop     chan struct{}
}

func (s *store) Informer() cache.SharedIndexInformer {
	return s.informer
}

func (s *store) AddIndexer(indexers cache.Indexers) error {
	return s.informer.AddIndexers(indexers)
}

func (s *store) Run() {
	go s.informer.Run(s.stop)
}

func (s *store) Down() {
	close(s.stop)
}

func (s *store) List() []interface{} {
	return s.informer.GetStore().List()
}

func (s *store) IndexFilter(ss string, index index.IndexName) ([]interface{}, error) {
	return s.informer.GetIndexer().ByIndex(string(index), ss)
}

func (s *store) Get(name string, namespace string) (interface{}, bool, error) {
	var key string
	if len(namespace) == 0 {
		key = name
	} else {
		key = fmt.Sprintf("%s/%s", namespace, name)
	}
	return s.informer.GetStore().GetByKey(key)
}
