package index

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
)

type IndexName string

const (
	IPIndex         IndexName = "ip"
	NodeNameIndex   IndexName = "nodeName"
	LabelIndex      IndexName = "label"
	ImageIndex      IndexName = "image"
	AnnotationIndex IndexName = "annotation"
	NamespaceIndex  IndexName = "namespace"
	NameIndex       IndexName = "name"
)

func isValidPod(obj interface{}) (*v1.Pod, error) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		return nil, fmt.Errorf("not implemented")
	}
	return pod, nil
}

func PodIPIndexer(obj interface{}) ([]string, error) {
	pod, err := isValidPod(obj)
	if err != nil {
		return nil, err
	}
	var ips []string
	for _, ip := range pod.Status.PodIPs {
		ips = append(ips, ip.IP)
	}
	return ips, nil
}

func LabelIndexer(obj interface{}) ([]string, error) {
	m, err := meta.Accessor(obj)
	if err != nil {
		return nil, err
	}
	labelMap := m.GetLabels()
	var labels []string
	for k, v := range labelMap {
		label := fmt.Sprintf("%s=%s", k, v)
		labels = append(labels, label)
	}
	return labels, nil
}

func AnnotationIndexer(obj interface{}) ([]string, error) {
	m, err := meta.Accessor(obj)
	if err != nil {
		return nil, err
	}
	annotationMap := m.GetAnnotations()
	var annotations []string
	for k, v := range annotationMap {
		annotation := fmt.Sprintf("%s=%s", k, v)
		annotations = append(annotations, annotation)
	}
	return annotations, nil
}

func PodImageIndexer(obj interface{}) ([]string, error) {
	var images []string
	for _, container := range obj.(*v1.Pod).Spec.Containers {
		images = append(images, container.Image)
	}
	return images, nil
}

func NameIndexer(obj interface{}) ([]string, error) {
	m, err := meta.Accessor(obj)
	if err != nil {
		return nil, err
	}
	return []string{m.GetName()}, nil
}

func PodNodeNameIndexer(obj interface{}) ([]string, error) {
	pod, err := isValidPod(obj)
	if err != nil {
		return nil, err
	}
	return []string{pod.Spec.NodeName}, nil
}
