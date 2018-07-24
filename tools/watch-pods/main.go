package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"

	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/kyma-project/kyma/tools/watch-pods/internal/tester"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	minWaitingPeriod := flag.Duration("minWaitingPeriod", time.Minute, "Minimum waiting period")
	maxWaitingPeriod := flag.Duration("maxWaitingPeriod", time.Minute*3, "Maximum waiting period")
	reqStabilityPeriod := flag.Duration("reqStabilityPeriod", time.Minute, "Required stability period")
	ignorePodsPattern := flag.String("ignorePodsPattern", "", "Regexp for pod name that containers will be ignored")
	ignoreNsPattern := flag.String("ignoreNsPattern", "", "Regexp for namespace that containers will be ignored")
	ignoreContainerPattern := flag.String("ignoreContainersPattern", "", "Regexp for namespaces that containers will be ignored")

	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	ignoreFunc, err := tester.IgnoreContainersByRegexp(*ignorePodsPattern, *ignoreNsPattern, *ignoreContainerPattern)
	if err != nil {
		panic(err.Error())
	}
	clusterState := tester.NewClusterState(ignoreFunc)
	lw := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "pods", "", fields.Everything())
	watcher := tester.NewPodWatcher(lw, createOnPodUpdateFunc(clusterState), createOnPodDeleteFunc(clusterState))
	if err := watcher.StartListeningToEvents(); err != nil {
		panic(err.Error())
	}

	deadline := time.Now().Add(*maxWaitingPeriod)

	fmt.Printf("Run tester with following configuration: min waiting: %v, max waiting: %v\n", minWaitingPeriod, maxWaitingPeriod)
	if *ignorePodsPattern != "" || *ignoreNsPattern != "" || *ignoreContainerPattern != "" {
		fmt.Printf("Ignore configuration: pod pattern: '%s', ns pattern: '%s', container pattern: '%s'", *ignorePodsPattern, *ignoreNsPattern, *ignoreContainerPattern)
	}

	stable := false
	var unstableContainers []tester.ContainerAndState

	<-time.After(*minWaitingPeriod)
	for time.Now().Before(deadline) {
		unstableContainers = clusterState.GetUnstableContainers(*reqStabilityPeriod)
		size := len(unstableContainers)
		fmt.Printf("Got %d unstable containers: %s \n", size, shortContainersDescription(unstableContainers))
		if size == 0 {
			stable = true
			break
		}
		<-time.After(time.Second * 10)
	}

	watcher.Stop()

	if !stable {
		fmt.Println("Unstable containers")
		for _, c := range unstableContainers {
			fmt.Printf("%+v\n", c)
		}
		os.Exit(1)

	}

}

func createOnPodUpdateFunc(clusterState *tester.ClusterState) func(p *v1.Pod) {
	return func(p *v1.Pod) {
		for _, c := range p.Status.ContainerStatuses {
			clusterState.UpdateState(tester.Container{Ns: p.Namespace, PodName: p.Name, ContainerName: c.Name}, tester.StateUpdate{
				Ready:      c.Ready,
				RestartCnt: c.RestartCount,
			})
		}
	}
}

func createOnPodDeleteFunc(clusterState *tester.ClusterState) func(ns, name string) {
	return func(ns, podName string) {
		clusterState.ForgetPod(ns, podName)
	}
}

func shortContainersDescription(containers []tester.ContainerAndState) string {
	out := ""
	for _, c := range containers {
		out += fmt.Sprintf("{Ns: %s, PodName: %s, Container: %s} ", c.Container.Ns, c.Container.PodName, c.Container.ContainerName)
	}
	return "[" + out + "]"
}
