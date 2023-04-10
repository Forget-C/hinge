package index

import (
	"k8s.io/client-go/tools/cache"
)

var Mapping = new(IndexerMapping)

const wildcard = "*"

func init() {
	Mapping.Register("pods", NodeNameIndex, PodNodeNameIndexer)
	Mapping.Register("pods", IPIndex, PodIPIndexer)
	Mapping.Register("pods", LabelIndex, LabelIndexer)
	Mapping.Register("pods", ImageIndex, PodImageIndexer)
	Mapping.Register("pods", AnnotationIndex, AnnotationIndexer)
	Mapping.Register("pods", NameIndex, NameIndexer)

	Mapping.Register("services", LabelIndex, LabelIndexer)
	Mapping.Register("services", AnnotationIndex, AnnotationIndexer)
	Mapping.Register("services", NameIndex, NameIndexer)
	
	Mapping.Register("nodes", LabelIndex, LabelIndexer)
	Mapping.Register("nodes", AnnotationIndex, AnnotationIndexer)
	Mapping.Register("nodes", NameIndex, NameIndexer)
}

type IndexerMapping struct {
	mapping map[string]map[IndexName]cache.IndexFunc
}

func (i *IndexerMapping) Register(resource string, name IndexName, f cache.IndexFunc) {
	if i.mapping == nil {
		i.mapping = make(map[string]map[IndexName]cache.IndexFunc)
	}
	if _, ok := i.mapping[resource]; !ok {
		i.mapping[resource] = make(map[IndexName]cache.IndexFunc)
	}
	i.mapping[resource][name] = f
}

func (i *IndexerMapping) Get(resource string, name ...IndexName) cache.Indexers {
	indexers := cache.Indexers{}
	for _, n := range name {
		if _, ok := i.mapping[resource]; ok {
			indexers[string(n)] = i.mapping[resource][n]
		} else {
			if _, ok := i.mapping[wildcard]; ok {
				indexers[string(n)] = i.mapping[wildcard][n]
			}
		}
	}
	return indexers
}
