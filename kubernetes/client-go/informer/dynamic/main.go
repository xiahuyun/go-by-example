package main

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

func main() {

	config, err := clientcmd.BuildConfigFromFlags("", "/Users/hxia/.kube/config")
	if err != nil {
		panic(err)
	}

	dyn := dynamic.NewForConfigOrDie(config)

	gvr := schema.GroupVersionResource{
		Group:    "webapp.my.domain",
		Version:  "v1",
		Resource: "guestbooks",
	}

	// DynamicInformerFactory：类似 SharedInformerFactory，但产出 dynamic informer
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(dyn, 0, metav1.NamespaceAll, nil)
	informer := factory.ForResource(gvr).Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			u := obj.(*unstructured.Unstructured)
			name := u.GetName()
			replicas, _, _ := unstructured.NestedInt64(u.Object, "spec", "replicas")
			fmt.Printf("ADDED %s replicas=%d\n", name, replicas)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			// ...
		},
		DeleteFunc: func(obj interface{}) {
			// DFSU 也会到这里，obj 可能是 DeletedFinalStateUnknown
		},
	})

	// ---------- 启动 ----------
	ctx := context.Background()
	go informer.Run(ctx.Done())

	// 等缓存同步（非常重要，否则 List 可能还没回来）
	if !cache.WaitForCacheSync(ctx.Done(), informer.HasSynced) {
		panic("timeout waiting for cache sync")
	}
	fmt.Println("Cache synced, ready to work")

	// 阻塞住，看事件
	<-ctx.Done()
}

func init() {
	utilruntime.PanicHandlers = []func(context.Context, interface{}){
		func(ctx context.Context, v interface{}) {
			// 使用 klog 记录 panic 信息，不退出进程
			klog.Errorf("Recovered from panic: %v", v)
		},
	}
	utilruntime.ReallyCrash = false
}
