/*
Code in this file is based on code from Kubernetes Helm.
Original code was licensed under the Apache License, Version 2.0 with copyright assigned to "2016 The Kubernetes Authors All rights reserved".
*/

package memory

import "sync"

type threadSafeStorage struct {
	sync.RWMutex
}

// lockW locks storage for writing
func (s *threadSafeStorage) lockW() func() {
	s.Lock()
	return func() { s.Unlock() }
}

// lockR locks storage for reading
func (s *threadSafeStorage) lockR() func() {
	s.RLock()
	return func() { s.RUnlock() }
}

// unlock calls fn which reverses a lockR or lockW. e.g:
// ```defer unlock(s.lockR())```, locks mem for reading at the
// call point of defer and unlocks upon exiting the block.
func unlock(fn func()) { fn() }
