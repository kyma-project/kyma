package tester

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateClusterState(t *testing.T) {
	t.Run("when first success update for Container", func(t *testing.T) {
		// GIVEN
		tp := timeProvider()
		cluster := NewClusterStateWithTimeProvider(tp)
		readyContainer := Container{Ns: "ns1", ContainerName: "cn1", PodName: "pn1"}
		readyState := StateUpdate{Ready: true, RestartCnt: 0}
		cluster.UpdateState(readyContainer, readyState)

		unreadyContainer := Container{Ns: "ns2", ContainerName: "cn2", PodName: "pn2"}
		unreadyState := StateUpdate{Ready: false, RestartCnt: 0}
		cluster.UpdateState(unreadyContainer, unreadyState)
		// WHEN
		actUnstable := cluster.GetUnstableContainers(time.Hour)
		// THEN
		assert.Len(t, actUnstable, 2)
		assert.Contains(t, actUnstable, ContainerAndState{Container: readyContainer, State: State{Ready: true, RestartCnt: 0, ReadySince: fixTime()}})
		assert.Contains(t, actUnstable, ContainerAndState{Container: unreadyContainer, State: State{Ready: false, RestartCnt: 0, ReadySince: time.Time{}}})

		// WHEN
		actUnstable = cluster.GetUnstableContainers(time.Second)
		assert.Len(t, actUnstable, 1)
		assert.Contains(t, actUnstable, ContainerAndState{Container: unreadyContainer, State: State{Ready: false, RestartCnt: 0, ReadySince: time.Time{}}})

	})

	t.Run("when ready become unready", func(t *testing.T) {
		// GIVEN
		tp := timeProvider()
		cluster := NewClusterStateWithTimeProvider(tp)
		container := Container{Ns: "ns1", ContainerName: "cn1", PodName: "pn1"}
		readyState := StateUpdate{Ready: true, RestartCnt: 0}
		cluster.UpdateState(container, readyState)

		unreadyState := StateUpdate{Ready: false, RestartCnt: 1}
		cluster.UpdateState(container, unreadyState)

		// WHEN
		actUnstable := cluster.GetUnstableContainers(time.Second)
		// THEN
		assert.Len(t, actUnstable, 1)
		assert.Contains(t, actUnstable, ContainerAndState{Container: container, State: State{Ready: false, RestartCnt: 1, ReadySince: time.Time{}}})

	})

	t.Run("when ready is still ready", func(t *testing.T) {
		// GIVEN
		tp := timeProvider()
		cluster := NewClusterStateWithTimeProvider(tp)
		container := Container{Ns: "ns1", ContainerName: "cn1", PodName: "pn1"}
		readyState := StateUpdate{Ready: true, RestartCnt: 0}
		cluster.UpdateState(container, readyState)
		cluster.UpdateState(container, readyState)
		// WHEN container is stable since ~ 1 second
		actUnstable := cluster.GetUnstableContainers(time.Hour)
		// THEN
		assert.Len(t, actUnstable, 1)
		assert.Contains(t, actUnstable, ContainerAndState{Container: container, State: State{Ready: true, RestartCnt: 0, ReadySince: fixTime()}})

		actUnstable = cluster.GetUnstableContainers(time.Second)
		assert.Empty(t, actUnstable)

	})

	t.Run("when becomes ready", func(t *testing.T) {
		// GIVEN
		tp := timeProvider()
		cluster := NewClusterStateWithTimeProvider(tp)
		container := Container{Ns: "ns1", ContainerName: "cn1", PodName: "pn1"}
		unreadyState := StateUpdate{Ready: false, RestartCnt: 1}
		cluster.UpdateState(container, unreadyState)

		readyState := StateUpdate{Ready: true, RestartCnt: 1}
		cluster.UpdateState(container, readyState)
		//WHEN
		actUnstable := cluster.GetUnstableContainers(time.Hour)
		// THEN
		assert.Len(t, actUnstable, 1)
		assert.Contains(t, actUnstable, ContainerAndState{Container: container, State: State{Ready: true, RestartCnt: 1, ReadySince: fixTime()}})
	})
}

func TestIgnoreUnstableContainers(t *testing.T) {
	// GIVEN
	tp := timeProvider()
	ignoreFunc, err := IgnoreContainersByRegexp("a", "a", "a")
	require.NoError(t, err)
	cs := NewClusterState(ignoreFunc)
	cs.timeProvider = tp
	unreadyContainer := Container{Ns: "a", ContainerName: "a", PodName: "a"}
	unreadyState := StateUpdate{Ready: false, RestartCnt: 123}

	cs.UpdateState(unreadyContainer, unreadyState)
	// WHEN
	actualUnstableContainers := cs.GetUnstableContainers(time.Hour)
	// THEN
	assert.Empty(t, actualUnstableContainers)
}

func TestIgnoreContainersByRegexp(t *testing.T) {
	ignoreFunc, err := IgnoreContainersByRegexp("dex-web-.*", "ysf-system", "dex-web")
	require.NoError(t, err)
	ignoredPod := "dex-web-84f6d87df8-tdvzf"
	ignoredNs := "ysf-system"
	ignoredContainer := "dex-web"

	t.Run("should ignore", func(t *testing.T) {
		toIgnore := Container{PodName: ignoredPod, Ns: ignoredNs, ContainerName: ignoredContainer}
		assert.True(t, ignoreFunc(toIgnore))
	})

	t.Run("should not ignore - different pod", func(t *testing.T) {
		toNotIgnore := Container{PodName: "sth", Ns: ignoredNs, ContainerName: ignoredContainer}
		assert.False(t, ignoreFunc(toNotIgnore))
	})

	t.Run("should not ignore - different ns", func(t *testing.T) {
		toNotIgnore := Container{PodName: ignoredPod, Ns: "sth", ContainerName: ignoredContainer}
		assert.False(t, ignoreFunc(toNotIgnore))

	})

	t.Run("should not ignore - different container", func(t *testing.T) {
		toNotIgnore := Container{PodName: ignoredPod, Ns: ignoredNs, ContainerName: "sth"}
		assert.False(t, ignoreFunc(toNotIgnore))

	})
}

func TestIgnoreContainersRegexpCombinations(t *testing.T) {
	aContainer := Container{PodName: "a", Ns: "a", ContainerName: "a"}
	bContainer := Container{PodName: "b", Ns: "b", ContainerName: "b"}

	t.Run("no patterns", func(t *testing.T) {
		ignoreFunc, err := IgnoreContainersByRegexp("", "", "")
		require.NoError(t, err)
		assert.False(t, ignoreFunc(aContainer))
		assert.False(t, ignoreFunc(bContainer))
	})

	t.Run("no pattern for pod", func(t *testing.T) {
		ignoreFunc, err := IgnoreContainersByRegexp("", "a", "a")
		require.NoError(t, err)
		assert.True(t, ignoreFunc(aContainer))
		assert.False(t, ignoreFunc(bContainer))
	})

	t.Run("no pattern for ns", func(t *testing.T) {
		ignoreFunc, err := IgnoreContainersByRegexp("a", "", "a")
		require.NoError(t, err)
		assert.True(t, ignoreFunc(aContainer))
		assert.False(t, ignoreFunc(bContainer))
	})

	t.Run("no pattern for container", func(t *testing.T) {
		ignoreFunc, err := IgnoreContainersByRegexp("a", "a", "")
		require.NoError(t, err)
		assert.True(t, ignoreFunc(aContainer))
		assert.False(t, ignoreFunc(bContainer))
	})

	t.Run("complex patterns", func(t *testing.T) {
		aContainer := Container{PodName: "a", Ns: "a", ContainerName: "a"}
		bContainer := Container{PodName: "b", Ns: "b", ContainerName: "b"}
		ignoreFunc, err := IgnoreContainersByRegexp("a|b", "", "")
		require.NoError(t, err)
		assert.True(t, ignoreFunc(aContainer))
		assert.True(t, ignoreFunc(bContainer))

	})

	givenInvalidPattern := "a(b"
	t.Run("invalid pod pattern", func(t *testing.T) {
		_, err := IgnoreContainersByRegexp(givenInvalidPattern, "", "")
		assert.Error(t, err)

	})

	t.Run("invalid ns pattern", func(t *testing.T) {
		_, err := IgnoreContainersByRegexp("", givenInvalidPattern, "")
		assert.Error(t, err)
	})

	t.Run("invalid container pattern", func(t *testing.T) {
		_, err := IgnoreContainersByRegexp("", "", givenInvalidPattern)
		assert.Error(t, err)
	})
}

// helpers

func NewClusterStateWithTimeProvider(tp func() time.Time) *ClusterState {
	cs := NewClusterState(nil)
	cs.timeProvider = tp
	return cs
}

func timeProvider() func() time.Time {
	i := 0
	return func() time.Time {
		defer func() {
			i++
		}()
		return fixTime().Add(time.Second * time.Duration(i))
	}
}

func fixTime() time.Time {
	return time.Date(2017, 11, 6, 0, 0, 0, 0, time.UTC)

}
