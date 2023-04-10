package hinge

import (
	"fmt"
	"github.com/Forget-C/hinge/filter"
	"github.com/Forget-C/hinge/index"
	"github.com/Forget-C/hinge/store"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"strings"
)

type Option func(cluster *Cluster)

func WithIndex(setting map[string][]index.IndexName) Option {
	return func(cluster *Cluster) {
		cluster.index = setting
	}
}

func DefaultIndexSetting() map[string][]index.IndexName {
	return map[string][]index.IndexName{
		"pods": []index.IndexName{
			index.NodeNameIndex,
			index.IPIndex,
			index.LabelIndex,
			index.ImageIndex,
			index.AnnotationIndex,
			index.NameIndex,
		},
		"services": []index.IndexName{
			index.LabelIndex,
			index.AnnotationIndex,
			index.NameIndex,
		},
		"nodes": []index.IndexName{
			index.LabelIndex,
			index.AnnotationIndex,
			index.NameIndex,
		},
	}
}

func NewCluster(id string, cli *kubernetes.Clientset, option ...Option) *Cluster {
	cluster := &Cluster{
		id:     id,
		client: cli,
	}
	for _, opt := range option {
		opt(cluster)
	}
	if cluster.index == nil {
		cluster.index = DefaultIndexSetting()
	}
	err := cluster.init()
	if err != nil {
		klog.Error(err)
	}
	return cluster
}

type Cluster struct {
	id                 string
	client             *kubernetes.Clientset
	index              map[string][]index.IndexName
	factory            informers.SharedInformerFactory
	preferredResources []*metav1.APIResourceList
	stores             map[string]store.Store
}

func (c *Cluster) init() error {
	serverResource, err := c.client.ServerPreferredResources()
	if err != nil {
		return err
	}
	c.preferredResources = serverResource
	c.factory = informers.NewSharedInformerFactory(c.client, 0)
	c.stores = make(map[string]store.Store)
	return nil
}

func (c *Cluster) getGVR(resource string) (schema.GroupVersionResource, error) {
	for _, r := range c.preferredResources {
		for _, api := range r.APIResources {
			if strings.EqualFold(api.Name, resource) {
				gv, err := schema.ParseGroupVersion(r.GroupVersion)
				if err != nil {
					return schema.GroupVersionResource{}, err
				}
				return schema.GroupVersionResource{
					Group:    gv.Group,
					Version:  gv.Version,
					Resource: resource,
				}, nil
			}
		}
	}
	return schema.GroupVersionResource{}, fmt.Errorf("unsupported resource: %s", resource)
}

func (c *Cluster) Start(stopCh <-chan struct{}) {
	for r, inxName := range c.index {
		gvr, err := c.getGVR(r)
		if err != nil {
			klog.Error(err)
			continue
		}
		informer, err := c.factory.ForResource(gvr)
		if err != nil {
			klog.Error(err)
			continue
		}
		indexers := index.Mapping.Get(r, inxName...)
		_ = informer.Informer().AddIndexers(indexers)
		c.stores[r] = store.NewStore(informer.Informer())
	}
	c.factory.Start(stopCh)
}

func (c *Cluster) WaitForCacheSync(stopCh <-chan struct{}) {
	c.factory.WaitForCacheSync(stopCh)
}

func (c *Cluster) Builder(resource string) filter.Builder {
	return filter.NewDefaultFilter(c.stores[resource])
}
