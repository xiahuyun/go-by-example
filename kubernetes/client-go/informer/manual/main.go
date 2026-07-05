package main

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

func main() {
	config, err := clientcmd.BuildConfigFromFlags("", "/Users/hxia/.kube/config")
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)

	// 2. 启动手动 ListAndWatch 循环
	manualListAndWatch(clientset)
}

func manualListAndWatch(clientset *kubernetes.Clientset) {
	// 设置超时 context（整个循环）
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// 用于 Watch 的字段选择器（可选）
	fieldSelector := fields.Everything()

	// 初始 resourceVersion 为空，由 List 返回
	var resourceVersion string

	for {
		select {
		case <-ctx.Done():
			klog.Info("Context cancelled, stopping ListAndWatch")
			return
		default:
		}

		// ---- Step 1: List ----
		var allPods []v1.Pod
		continueToken := ""
		for {
			listOptions := metav1.ListOptions{
				FieldSelector: fieldSelector.String(),
				Limit:         500, // 分页，避免一次拉取太多
				Continue:      continueToken,
			}
			podList, err := clientset.CoreV1().Pods(metav1.NamespaceAll).List(ctx, listOptions)
			if err != nil {
				klog.Errorf("List failed: %v, retrying in 5 seconds...", err)
				time.Sleep(5 * time.Second)
				continue
			}
			allPods = append(allPods, podList.Items...)
			continueToken = podList.Continue
			if continueToken == "" {
				resourceVersion = podList.ResourceVersion
				break
			}
		}

		// 处理 List 返回的所有 Pod（相当于全量同步）
		for _, pod := range allPods {
			processPodEvent(watch.Added, &pod)
		}

		// 获取最新的 resourceVersion，用于后续 Watch
		klog.Infof("List completed, resourceVersion=%s, %d pods", resourceVersion, len(allPods))

		// ---- Step 2: Watch ----
		// 设置 Watch 超时（例如 30 秒后自动断开，重新 List 以同步）
		// 这里我们可以不设置超时，而是一直 watch 直到错误才重新 List 吗？
		// client-go 是怎么做的？
		watchTimeout := int64(300)
		watchOptions := metav1.ListOptions{
			ResourceVersion: resourceVersion,
			FieldSelector:   fieldSelector.String(),
			TimeoutSeconds:  &watchTimeout,
		}

		watcher, err := clientset.CoreV1().Pods(metav1.NamespaceAll).Watch(ctx, watchOptions)
		if err != nil {
			klog.Errorf("Watch failed: %v, will re-list after 5 seconds", err)
			time.Sleep(5 * time.Second)
			continue
		}

		// 处理 Watch 事件
		watchDone := make(chan struct{})
		go func() {
			defer close(watchDone)
			for event := range watcher.ResultChan() {
				pod, ok := event.Object.(*v1.Pod)
				if !ok {
					klog.Warningf("Received non-Pod object: %T", event.Object)
					continue
				}
				processPodEvent(event.Type, pod)
			}
		}()

		// 等待 Watch 结束（超时或错误）
		<-watchDone
		klog.Info("Watch ended, will re-list to ensure consistency")
		// 循环回到 List，重新同步
	}
}

// processPodEvent 处理 Pod 事件（示例：打印）
func processPodEvent(eventType watch.EventType, pod *v1.Pod) {
	switch eventType {
	case watch.Added:
		fmt.Printf("ADDED   %s/%s/%s\n", pod.Namespace, pod.Name, pod.ResourceVersion)
	case watch.Modified:
		fmt.Printf("MODIFIED %s/%s\n", pod.Namespace, pod.Name)
	case watch.Deleted:
		fmt.Printf("DELETED %s/%s\n", pod.Namespace, pod.Name)
	case watch.Error:
		fmt.Printf("ERROR event for pod %s/%s\n", pod.Namespace, pod.Name)
	}
}
