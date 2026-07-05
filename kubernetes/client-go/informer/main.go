package main

import (
	"log"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	config, err := clientcmd.BuildConfigFromFlags("", "/Users/hxia/.kube/config")
	if err != nil {
		panic(err)
	}

	// 为什么需要创建一个clientset,其它 client 不行吗？
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	stopCh := make(chan struct{})
	defer close(stopCh)

	sharedInformerFactory := informers.NewSharedInformerFactory(clientset, time.Minute)
	podInformer := sharedInformerFactory.Core().V1().Pods().Informer()

	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			mObj := obj.(v1.Object)
			log.Printf("New Pod Added to Store: %s", mObj.GetName())
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oObj := oldObj.(v1.Object)
			nObj := newObj.(v1.Object)
			log.Printf("%s Pod Updated to %s", oObj.GetName(), nObj.GetName())
		},
		DeleteFunc: func(obj interface{}) {
			mObj := obj.(v1.Object)
			log.Printf("Pod Deleted from Store: %s", mObj.GetName())
		},
	})

	podInformer.Run(stopCh)
}
