package apps

import (
	"sync"
)

// Name is the name of an app that was created by a test in the suite. It is a fmt.Stringer
type Name string

// NameFromString creates a new Name from a string representation of an app name
func NameFromString(str string) Name {
	return Name(str)
}

// String is the fmt.Stringer interface implementation
func (a Name) String() string {
	return string(a)
}

// Set is a concurrency-safe set of app names
type Set struct {
	rwm *sync.RWMutex
	set map[Name]struct{}
}

// NewSet creates a new, empty set of AppNames
func NewSet() *Set {
	return &Set{rwm: new(sync.RWMutex), set: make(map[Name]struct{})}
}

// Add adds appName to the set and returns whether or not the app name was already in the set
func (s *Set) Add(appName Name) bool {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	_, ok := s.set[appName]
	s.set[appName] = struct{}{}
	return !ok
}

// GetAll returns all app names in the set. Note that more app names can be added to or removed from the set after this call returns the set after this call returns.
func (s *Set) GetAll() []Name {
	s.rwm.RLock()
	s.rwm.RUnlock()
	ret := make([]Name, len(s.set))
	i := 0
	for appName := range s.set {
		ret[i] = appName
		i++
	}
	return ret
}

// Clear clears all app names from the set and returns how many were in it
func (s *Set) Clear() int {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	n := len(s.set)
	s.set = make(map[Name]struct{})
	return n
}
