package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const (
	defaultNamespace = "informer-410-demo"
	demoLabelKey     = "app.kubernetes.io/name"
	demoLabelValue   = "informer-410-gone-real"
)

type options struct {
	kubeconfig     string
	namespace      string
	mode           string
	resourceVer    string
	churn          int
	cleanup        bool
	timeout        time.Duration
	watchTimeout   time.Duration
	informerRunFor time.Duration
}

func main() {
	opts := parseFlags()

	ctx, cancel := context.WithTimeout(context.Background(), opts.timeout)
	defer cancel()

	clientset, err := newClientset(opts.kubeconfig)
	if err != nil {
		log.Fatalf("build kubernetes client: %v", err)
	}

	if err := ensureNamespace(ctx, clientset, opts.namespace); err != nil {
		log.Fatalf("ensure namespace: %v", err)
	}

	if opts.cleanup {
		defer func() {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := cleanupDemoConfigMaps(cleanupCtx, clientset, opts.namespace); err != nil {
				log.Printf("[cleanup] failed: %v", err)
			}
		}()
	}

	switch opts.mode {
	case "manual":
		err = runManual(ctx, clientset, opts)
	case "informer":
		err = runInformer(ctx, clientset, opts)
	default:
		err = fmt.Errorf("unknown mode %q, want manual or informer", opts.mode)
	}
	if err != nil {
		log.Fatalf("demo failed: %v", err)
	}
}

func parseFlags() options {
	homeKubeconfig := ""
	if home := homedir.HomeDir(); home != "" {
		homeKubeconfig = filepath.Join(home, ".kube", "config")
	}

	opts := options{}
	flag.StringVar(&opts.kubeconfig, "kubeconfig", homeKubeconfig, "path to kubeconfig")
	flag.StringVar(&opts.namespace, "namespace", defaultNamespace, "namespace used by the demo")
	flag.StringVar(&opts.mode, "mode", "manual", "manual or informer")
	flag.StringVar(&opts.resourceVer, "resource-version", "", "stale resourceVersion to watch from; empty means capture one before churn")
	flag.IntVar(&opts.churn, "churn", 3000, "number of configmap create/update/delete operations used to advance resourceVersion")
	flag.BoolVar(&opts.cleanup, "cleanup", true, "delete demo configmaps at exit")
	flag.DurationVar(&opts.timeout, "timeout", 3*time.Minute, "overall demo timeout")
	flag.DurationVar(&opts.watchTimeout, "watch-timeout", 20*time.Second, "single watch timeout")
	flag.DurationVar(&opts.informerRunFor, "informer-run-for", 45*time.Second, "how long informer mode keeps running after it starts")
	flag.Parse()
	return opts
}

func newClientset(kubeconfig string) (*kubernetes.Clientset, error) {
	if kubeconfig == "" {
		return nil, errors.New("kubeconfig is empty")
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func runManual(ctx context.Context, clientset *kubernetes.Clientset, opts options) error {
	staleRV, err := staleResourceVersion(ctx, clientset, opts)
	if err != nil {
		return err
	}

	log.Printf("[manual] watch configmaps from stale resourceVersion=%s", staleRV)
	gone, err := watchOnce(ctx, clientset, opts.namespace, staleRV, opts.watchTimeout)
	if err != nil {
		return err
	}
	if !gone {
		log.Printf("[manual] apiserver still accepts resourceVersion=%s", staleRV)
		log.Printf("[manual] try --resource-version=1 or increase --churn if you want to force 410 Gone")
		return nil
	}

	log.Printf("[manual] detected 410 Gone, recover by relisting")
	list, err := listDemoConfigMaps(ctx, clientset, opts.namespace)
	if err != nil {
		return fmt.Errorf("relist after 410 Gone: %w", err)
	}
	log.Printf("[manual] relist complete: objects=%d newResourceVersion=%s", len(list.Items), list.ResourceVersion)

	log.Printf("[manual] restart watch from fresh resourceVersion=%s", list.ResourceVersion)
	_, err = watchOnce(ctx, clientset, opts.namespace, list.ResourceVersion, 5*time.Second)
	return err
}

func runInformer(ctx context.Context, clientset *kubernetes.Clientset, opts options) error {
	staleRV, err := staleResourceVersion(ctx, clientset, opts)
	if err != nil {
		return err
	}

	var watchCalls int32
	lw := &cache.ListWatch{
		ListFunc: func(listOptions metav1.ListOptions) (runtime.Object, error) {
			listOptions.LabelSelector = demoLabelSelector()
			list, err := clientset.CoreV1().ConfigMaps(opts.namespace).List(ctx, listOptions)
			if err != nil {
				log.Printf("[informer:list] error: %v", err)
				return nil, err
			}
			log.Printf("[informer:list] objects=%d resourceVersion=%s", len(list.Items), list.ResourceVersion)
			return list, nil
		},
		WatchFunc: func(listOptions metav1.ListOptions) (watch.Interface, error) {
			call := atomic.AddInt32(&watchCalls, 1)
			listOptions.LabelSelector = demoLabelSelector()
			listOptions.TimeoutSeconds = ptrTo[int64](int64(opts.watchTimeout.Seconds()))
			if call == 1 {
				log.Printf("[informer:watch] first watch deliberately uses stale resourceVersion=%s", staleRV)
				listOptions.ResourceVersion = staleRV
			} else {
				log.Printf("[informer:watch] watch #%d uses reflector resourceVersion=%s", call, listOptions.ResourceVersion)
			}
			watcher, err := clientset.CoreV1().ConfigMaps(opts.namespace).Watch(ctx, listOptions)
			if isGoneError(err) {
				log.Printf("[informer:watch] apiserver returned 410 Gone to Reflector: %v", err)
			}
			return watcher, err
		},
	}

	informer := cache.NewSharedIndexInformer(
		lw,
		&corev1.ConfigMap{},
		0,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)
	_, err = informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			cm := obj.(*corev1.ConfigMap)
			log.Printf("[handler] ADD %s/%s rv=%s", cm.Namespace, cm.Name, cm.ResourceVersion)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			cm := newObj.(*corev1.ConfigMap)
			log.Printf("[handler] UPDATE %s/%s rv=%s", cm.Namespace, cm.Name, cm.ResourceVersion)
		},
		DeleteFunc: func(obj interface{}) {
			cm, ok := obj.(*corev1.ConfigMap)
			if !ok {
				tombstone := obj.(cache.DeletedFinalStateUnknown)
				cm = tombstone.Obj.(*corev1.ConfigMap)
			}
			log.Printf("[handler] DELETE %s/%s rv=%s", cm.Namespace, cm.Name, cm.ResourceVersion)
		},
	})
	if err != nil {
		return fmt.Errorf("add event handler: %w", err)
	}

	runCtx, cancel := context.WithTimeout(ctx, opts.informerRunFor)
	defer cancel()

	log.Printf("[informer] start; Reflector should recover after the first stale watch gets 410 Gone")
	go informer.Run(runCtx.Done())
	if !cache.WaitForCacheSync(runCtx.Done(), informer.HasSynced) {
		return errors.New("informer cache did not sync")
	}

	if err := createOrUpdateConfigMap(runCtx, clientset, opts.namespace, "informer-after-recover", "after-recover"); err != nil {
		return err
	}
	<-runCtx.Done()
	log.Printf("[informer] stopped after %s", opts.informerRunFor)
	return nil
}

func staleResourceVersion(ctx context.Context, clientset *kubernetes.Clientset, opts options) (string, error) {
	if opts.resourceVer != "" {
		log.Printf("[setup] use provided stale resourceVersion=%s", opts.resourceVer)
		return opts.resourceVer, nil
	}

	list, err := listDemoConfigMaps(ctx, clientset, opts.namespace)
	if err != nil {
		return "", fmt.Errorf("capture initial resourceVersion: %w", err)
	}

	staleRV := list.ResourceVersion
	log.Printf("[setup] captured initial resourceVersion=%s", staleRV)

	if opts.churn <= 0 {
		return staleRV, nil
	}
	if err := churnConfigMaps(ctx, clientset, opts.namespace, opts.churn); err != nil {
		return "", err
	}
	return staleRV, nil
}

func watchOnce(ctx context.Context, clientset *kubernetes.Clientset, namespace, resourceVersion string, timeout time.Duration) (bool, error) {
	timeoutSeconds := int64(timeout.Seconds())
	watcher, err := clientset.CoreV1().ConfigMaps(namespace).Watch(ctx, metav1.ListOptions{
		LabelSelector:       demoLabelSelector(),
		FieldSelector:       fields.Everything().String(),
		ResourceVersion:     resourceVersion,
		TimeoutSeconds:      &timeoutSeconds,
		AllowWatchBookmarks: true,
	})
	if isGoneError(err) {
		log.Printf("[watch] Watch() returned 410 Gone: %v", err)
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("start watch: %w", err)
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case event, ok := <-watcher.ResultChan():
			if !ok {
				log.Printf("[watch] result channel closed")
				return false, nil
			}
			if event.Type == watch.Error {
				if isGoneObject(event.Object) {
					log.Printf("[watch] received 410 Gone error event: %s", statusMessage(event.Object))
					return true, nil
				}
				return false, fmt.Errorf("watch error event: %s", statusMessage(event.Object))
			}
			log.Printf("[watch] %s %s", event.Type, objectName(event.Object))
		}
	}
}

func churnConfigMaps(ctx context.Context, clientset *kubernetes.Clientset, namespace string, count int) error {
	log.Printf("[churn] create/update/delete %d configmap revisions", count)
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("rv-churn-%05d", i)
		if err := createOrUpdateConfigMap(ctx, clientset, namespace, name, fmt.Sprintf("%d", i)); err != nil {
			return fmt.Errorf("create/update configmap %s: %w", name, err)
		}
		if i%2 == 1 {
			err := clientset.CoreV1().ConfigMaps(namespace).Delete(ctx, name, metav1.DeleteOptions{})
			if err != nil && !apierrors.IsNotFound(err) {
				return fmt.Errorf("delete configmap %s: %w", name, err)
			}
		}
		if (i+1)%500 == 0 {
			log.Printf("[churn] progressed %d/%d", i+1, count)
		}
	}
	log.Printf("[churn] done")
	return nil
}

func createOrUpdateConfigMap(ctx context.Context, clientset *kubernetes.Clientset, namespace, name, value string) error {
	cms := clientset.CoreV1().ConfigMaps(namespace)
	existing, err := cms.Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = cms.Create(ctx, &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				Labels: map[string]string{
					demoLabelKey: demoLabelValue,
				},
			},
			Data: map[string]string{"value": value},
		}, metav1.CreateOptions{})
		return err
	}
	if err != nil {
		return err
	}
	if existing.Labels == nil {
		existing.Labels = map[string]string{}
	}
	existing.Labels[demoLabelKey] = demoLabelValue
	if existing.Data == nil {
		existing.Data = map[string]string{}
	}
	existing.Data["value"] = value
	_, err = cms.Update(ctx, existing, metav1.UpdateOptions{})
	return err
}

func ensureNamespace(ctx context.Context, clientset *kubernetes.Clientset, namespace string) error {
	_, err := clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err == nil {
		log.Printf("[setup] namespace exists: %s", namespace)
		return nil
	}
	if !apierrors.IsNotFound(err) {
		return err
	}
	_, err = clientset.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: namespace},
	}, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	log.Printf("[setup] namespace created: %s", namespace)
	return nil
}

func listDemoConfigMaps(ctx context.Context, clientset *kubernetes.Clientset, namespace string) (*corev1.ConfigMapList, error) {
	return clientset.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: demoLabelSelector(),
	})
}

func cleanupDemoConfigMaps(ctx context.Context, clientset *kubernetes.Clientset, namespace string) error {
	log.Printf("[cleanup] delete demo configmaps in namespace=%s", namespace)
	return clientset.CoreV1().ConfigMaps(namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: demoLabelSelector(),
	})
}

func isGoneError(err error) bool {
	return err != nil && apierrors.IsGone(err)
}

func isGoneObject(obj runtime.Object) bool {
	status, ok := obj.(*metav1.Status)
	if !ok {
		return false
	}
	return status.Code == http.StatusGone || status.Reason == metav1.StatusReasonExpired
}

func statusMessage(obj runtime.Object) string {
	status, ok := obj.(*metav1.Status)
	if !ok {
		return fmt.Sprintf("%T", obj)
	}
	return fmt.Sprintf("code=%d reason=%s message=%q", status.Code, status.Reason, status.Message)
}

func objectName(obj runtime.Object) string {
	metaObj, ok := obj.(metav1.Object)
	if !ok {
		return fmt.Sprintf("%T", obj)
	}
	return fmt.Sprintf("%s/%s rv=%s", metaObj.GetNamespace(), metaObj.GetName(), metaObj.GetResourceVersion())
}

func demoLabelSelector() string {
	return fmt.Sprintf("%s=%s", demoLabelKey, demoLabelValue)
}

func ptrTo[T any](v T) *T {
	return &v
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.SetOutput(os.Stdout)
}
