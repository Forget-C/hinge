package tests

import (
	"github.com/Forget-C/hinge"
	"github.com/Forget-C/hinge/index"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"testing"
)

func TestA(t *testing.T) {
	cfg, err := kubernetes.NewForConfig(&rest.Config{
		Host: "127.0.0.1:54948",
		TLSClientConfig: rest.TLSClientConfig{
			CertFile: "/Users/chenyang/.minikube/profiles/minikube/client.crt",
			KeyFile:  "/Users/chenyang/.minikube/profiles/minikube/client.key",
			CAFile:   "/Users/chenyang/.minikube/ca.crt",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	cls := hinge.NewCluster("123", cfg, hinge.WithIndex(map[string][]index.IndexName{
		"pods": []index.IndexName{
			index.ImageIndex,
		},
	}))
	c := make(chan struct{})
	cls.Start(c)
	cls.WaitForCacheSync(c)
	var v []interface{}
	//total, err := cls.Builder("pods").Where(index.LabelIndex, "app=nginx").
	//	Where(index.NamespaceIndex, "default").Where(index.NameIndex, "nginx-1").WhereField([]string{"Spec.Containers.Image"}, "ngi", false).Find(&v)
	total, err := cls.Builder("pods").Where(index.ImageIndex, "nginx").Limit(10).Page(1).Find(&v)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(total, v)
	t.Log(v[0])
	t.Log(v[1])
	t.Log(v[2])
	t.Log(v[3])

}
