# hinge
这是一个使用informer的示例项目, 用于展示如何简单使用informer以及自定义indexer。
这个项目支持你用类似于orm的方式在集群中查找资源，这依赖于informer的indexer机制。

# 使用
1. 创建集群对象
```go
    // 以证书的方式连接集群
	cfg, err := kubernetes.NewForConfig(&rest.Config{
		Host: "127.0.0.1:6443",
		TLSClientConfig: rest.TLSClientConfig{
			CertFile: "client.crt",
			KeyFile:  "client.key",
			CAFile:   "ca.crt",
		},
	})
	if err != nil {
		panic(err)
	}
    // 生成集群对象，并为其设置id标识
	cls := hinge.NewCluster("my-cluster", cfg)
	c := make(chan struct{})
    // 启动（实际是启动informer）
	cls.Start(c)
    // 等待首次数据同步
	cls.WaitForCacheSync(c)
```
2. 查询
```go
var v []interface{}
// 所有条件都是and关系
total, err := cls.
    // 在pods资源中查找
    Builder("pods").
    // labels 包含 app=nginx
    Where(index.LabelIndex, "app=nginx").
    // namespace 为 default
	Where(index.NamespaceIndex, "default").
    // name 为 nginx-1
    Where(index.NameIndex, "nginx-1").
    // Spec.Containers.Image 字段的值为 nginx
    WhereField([]string{"Spec.Containers.Image"}, "nginx", true).
    Find(&v)
total, err = cls.
    // 在pods资源中查找
    Builder("pods").
    // Spec.Containers.Image 字段的值包含 ngi
    WhereField([]string{"Spec.Containers.Image"}, "ngi", false).
    Find(&v)
total, err = cls.
    // 在pods资源中查找
    Builder("pods").
    // image 为nginx
    Where(index.ImageIndex, "nginx").
    // 使用 ResourceVersion 字段排序。
    // 注意 这里的排序字段需要参考对应的结构体。
    // 使用kubectl 命令查看yaml， 字段为 metadata.resourceVersion
    // 结构体中， metadata为嵌入字段，所以是 ResourceVersion， 而不是 metadata.resourceVersion
    Sort("ResourceVersion").Find(&v)
total, err = cls.
    // 在pods资源中查找
    Builder("pods").
    // image 为nginx
    Where(index.ImageIndex, "nginx").
    // 取10条
    Limit(10).
    // 在第一页
    // 这里用的是 page而不是offset
    Page(1).
    Find(&v)
```
4. 设置indexer
informer中的索引维护是通过map完成的，虽然单个占用的空间不多，但是大批量时还是会有内存压力。
默认情况下，创建`cluster`会监听这些`resource`及`indexer`
```go
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
//可以通过`WithIndex`选项自定义
cls := hinge.NewCluster("123", cfg, hinge.WithIndex(map[string][]index.IndexName{
	"pods": []index.IndexName{
		index.ImageIndex,
	},
	}))
```
也可以注册自己的index
```go
func NameIndexer(obj interface{}) ([]string, error) {
	m, err := meta.Accessor(obj)
	if err != nil {
		return nil, err
	}
	return []string{m.GetName()}, nil
}
Mapping.Register("pods", "name_index", NameIndexer)
settings := DefaultIndexSetting()
settings[pods] = append(settings[pods],"name_index")
cls := hinge.NewCluster("123", cfg, hinge.WithIndex(settings))
```
# 原理
如上文所述，基础的查询功能利用了infomer的indexer机制，在此基础上做了一些封装。
通过索引名称匹配，完全利用了index的功能。
而指定字段查找和排序则用了反射。

# 性能
相同的功能代码，在公司的生产环境中以平稳运行半年。
监听的资源数量大概在10w左右， 内存占用2g～4g。
具体的性能报告就不提供了， 因为这个项目目前仅想给大家做为参考使用