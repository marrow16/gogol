package settings

import (
	"maps"
	"slices"
	"sync"
)

type Variables struct {
	mutex         sync.RWMutex
	m             map[string]string
	notifications []func(name string, del bool, all map[string]string)
}

func newVariables() *Variables {
	return &Variables{
		m: make(map[string]string),
	}
}

func (v *Variables) NotifyChanges(fn func(name string, del bool, all map[string]string)) {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	if fn != nil {
		v.notifications = append(v.notifications, fn)
	}
}

func notifyChanges(fns []func(name string, del bool, all map[string]string), name string, del bool, all map[string]string) {
	for _, fn := range fns {
		fn(name, del, all)
	}
}

func (v *Variables) SetAll(m map[string]string) {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	v.m = maps.Clone(m)
}

func (v *Variables) Set(key, value string) {
	v.mutex.Lock()
	v.m[key] = value
	fns := slices.Clone(v.notifications)
	all := maps.Clone(v.m)
	v.mutex.Unlock()
	notifyChanges(fns, key, false, all)
}

func (v *Variables) Get(key string) (string, bool) {
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	value, ok := v.m[key]
	return value, ok
}

func (v *Variables) Delete(key string) {
	v.mutex.Lock()
	delete(v.m, key)
	fns := slices.Clone(v.notifications)
	all := maps.Clone(v.m)
	v.mutex.Unlock()
	notifyChanges(fns, key, true, all)
}

func (v *Variables) DeleteAll() {
	v.mutex.Lock()
	v.m = make(map[string]string)
	fns := slices.Clone(v.notifications)
	all := maps.Clone(v.m)
	v.mutex.Unlock()
	notifyChanges(fns, "", true, all)
}

func (v *Variables) Clone() map[string]string {
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	return maps.Clone(v.m)
}
