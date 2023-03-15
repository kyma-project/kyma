package watcher

import (
	"fmt"
	"reflect"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const (
	defaultSize         = 10
	defaultResync       = time.Minute
	fieldSelectorFormat = "metadata.name=%s"
)

type Watcher struct {
	client               kubernetes.Interface
	namespace            string
	name                 string
	updateNotifiableList []UpdateNotifiable
}

func NewWatcher(client kubernetes.Interface, namespace, name string) *Watcher {
	return &Watcher{
		client:               client,
		namespace:            namespace,
		name:                 name,
		updateNotifiableList: make([]UpdateNotifiable, 0, defaultSize),
	}
}

func (w *Watcher) Watch() *Watcher {
	// this needs to be figured out !!
	defer runtime.HandleCrash()
	factory := informers.NewSharedInformerFactoryWithOptions(
		w.client,
		defaultResync,
		informers.WithNamespace(w.namespace),
		informers.WithTweakListOptions(func(o *metav1.ListOptions) {
			o.FieldSelector = fmt.Sprintf(fieldSelectorFormat, w.name)
		}),
	)
	configMapsInformer := factory.Core().V1().ConfigMaps().Informer()
	configMapsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: w.updateFunc,
	})
	// this needs to be figured out !!
	factory.Start(wait.NeverStop)
	factory.WaitForCacheSync(wait.NeverStop)
	return w
}

func (w *Watcher) OnUpdateNotify(notifiable UpdateNotifiable) *Watcher {
	w.updateNotifiableList = append(w.updateNotifiableList, notifiable)
	return w
}

func (w *Watcher) updateFunc(o interface{}, n interface{}) {
	var (
		ok    bool
		oldCM *corev1.ConfigMap
		newCM *corev1.ConfigMap
	)

	if oldCM, ok = o.(*corev1.ConfigMap); !ok {
		// figure out something for the logger
		//logger.LogIfError(fmt.Errorf("cannot convert old object to configmap"))
		return
	}
	if newCM, ok = n.(*corev1.ConfigMap); !ok {
		//logger.LogIfError(fmt.Errorf("cannot convert new object to configmap"))
		return
	}

	if !reflect.DeepEqual(oldCM.Data, newCM.Data) {
		for _, n := range w.updateNotifiableList {
			n.NotifyUpdate(newCM)
		}
	}
}
