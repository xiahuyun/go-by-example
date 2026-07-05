package main

import (
	"context"
	"fmt"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	config, err := clientcmd.BuildConfigFromFlags("", "/Users/hxia/.kube/config")
	if err != nil {
		panic(err)
	}

	config.APIPath = "api"
	config.GroupVersion = &corev1.SchemeGroupVersion
	config.NegotiatedSerializer = scheme.Codecs

	restClient, err := rest.RESTClientFor(config)
	if err != nil {
		panic(err)
	}

	result := &corev1.PodList{}
	// API: http://127.0.0.1:8080/api/v1/namespaces/default/pods?limit=500
	if err := restClient.Get().
		Namespace("kube-system").
		Resource("pods").
		VersionedParams(&metav1.ListOptions{Limit: 5}, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(result); err != nil {
		panic(err)
	}

	for _, d := range result.Items {
		fmt.Printf("NAMESPACE: %v \t NAME: %v \t STATUS:%+v\n", d.Namespace, d.Name, d.Status.Phase)
	}
}
