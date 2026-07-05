package main

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

func main() {

	config, err := clientcmd.BuildConfigFromFlags("", "/Users/hxia/.kube/config")
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	// ---------- 关键：自己构造 ListWatch ----------
	listWatcher := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return clientset.CoreV1().Pods(metav1.NamespaceAll).List(context.Background(), options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			// 这里可以用 fieldSelector / labelSelector
			options.FieldSelector = fields.Everything().String()
			return clientset.CoreV1().Pods(metav1.NamespaceAll).Watch(context.Background(), options)
		},
	}

	// ---------- NewSharedIndexInformer ----------
	// 参数：ListWatch, 对象类型, 同步周期(Resync), Indexers
	informer := cache.NewSharedIndexInformer(
		listWatcher,
		&v1.Pod{},
		0, // Resync 周期（0 = 不 resync，只靠 Watch）
		cache.Indexers{
			cache.NamespaceIndex: cache.MetaNamespaceIndexFunc, // 按 namespace 索引
		},
	)

	// ---------- 注册 Handlers（这就是“Shared”的由来：一个 Informer 多个 Handler） ----------
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*v1.Pod)
			fmt.Printf("[ADD] %s/%s\n", pod.Namespace, pod.Name)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldPod := oldObj.(*v1.Pod)
			newPod := newObj.(*v1.Pod)
			fmt.Printf("[UPDATE] %s/%s Phase: %s -> %s\n",
				newPod.Namespace, newPod.Name, oldPod.Status.Phase, newPod.Status.Phase)
		},
		DeleteFunc: func(obj interface{}) {
			pod := obj.(*v1.Pod)
			fmt.Printf("[DELETE] %s/%s\n", pod.Namespace, pod.Name)
		},
	})

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*v1.Pod)
			fmt.Printf("[ADD] pod resourceversion: %s\n", pod.ResourceVersion)
			//panic("not implemented")
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldPod := oldObj.(*v1.Pod)
			newPod := newObj.(*v1.Pod)
			fmt.Printf("[UPDATE] pod resourceversion: %s/%s Phase: %s -> %s\n",
				newPod.Namespace, newPod.Name, oldPod.ResourceVersion, newPod.ResourceVersion)
		},
		DeleteFunc: func(obj interface{}) {
			pod := obj.(*v1.Pod)
			fmt.Printf("[DELETE] pod resourceversion: %s\n", pod.ResourceVersion)
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

	// 可以从 Store 里查
	if obj, exists, err := informer.GetStore().GetByKey("kube-system/etcd"); err == nil && exists {
		pod := obj.(*v1.Pod)
		fmt.Printf("Found from store: %s Phase=%s\n", pod.Name, pod.Status.Phase)
	}

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
