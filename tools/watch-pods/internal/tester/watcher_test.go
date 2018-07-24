package tester

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktesting "k8s.io/client-go/tools/cache/testing"
)

func TestPodWatcherReactsOnUpdates(t *testing.T) {
	// GIVEN:
	fixPodA := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"}}
	fixPodB := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod2", Namespace: "default"}}

	source := ktesting.NewFakeControllerSource()

	gotPodA := make(chan struct{})
	gotPodB := make(chan struct{})

	watcher := NewPodWatcher(source,
		func(pod *v1.Pod) {
			switch pod.Name {
			case fixPodA.GetName():
				close(gotPodA)
			case fixPodB.GetName():
				close(gotPodB)
			}
		},
		func(ns, name string) {},
	)

	require.NoError(t, watcher.StartListeningToEvents())

	// WHEN:
	source.Add(fixPodA)
	source.Modify(fixPodB)

	// THEN:
	assertSignalReceived(t, gotPodA)
	assertSignalReceived(t, gotPodB)

	watcher.Stop()

}

func TestPodWatcherReactsOnDelete(t *testing.T) {
	// GIVEN:
	fixPodA := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"}}

	source := ktesting.NewFakeControllerSource()

	podDeleteCalled := make(chan struct{})

	watcher := NewPodWatcher(source,
		func(pod *v1.Pod) {},
		func(ns, name string) {
			// THEN:
			assert.Equal(t, fixPodA.GetNamespace(), ns)
			assert.Equal(t, fixPodA.GetName(), name)
			close(podDeleteCalled)
		},
	)
	require.NoError(t, watcher.StartListeningToEvents())

	source.Add(fixPodA)

	// WHEN:
	source.Delete(fixPodA)

	// THEN:
	assertSignalReceived(t, podDeleteCalled)

	watcher.Stop()
}

func assertSignalReceived(t *testing.T, c chan struct{}) bool {
	t.Helper()
	select {
	case <-time.After(time.Second):
		t.Error("timeout triggered")
		return false
	case <-c:
		return true
	}
}
