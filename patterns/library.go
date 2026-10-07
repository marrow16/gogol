package patterns

import (
	"maps"
	"strings"
	"sync"
)

func defaultPatternsLib() map[string]Pattern {
	const (
		o = false
		x = true
	)
	return map[string]Pattern{
		"Lucas Exploding Glider Gun": MustNewPatternFromRle(strings.NewReader(`#N Lucas Exploding Glider Gun
#O Lucas & Martin Rowlinson, November 2017
#C Copied from GoGoL (https://github.com/marrow16/gogol)
x = 5, y = 5, rule = B0138/S124
4o$2o2bo$ob3o$ob2o$b2o!`)),
		"GoGoL": MustNewPatternFromRle(strings.NewReader(`#N GoGoL
#O Marrow, 2026
#C Default built-in pattern for GoGoL
#C https://github.com/marrow16/gogol
x = 28, y = 7, rule = B3/S23
b3o9b3o8bo$o3bo7bo3bo7bo$o6b3o2bo6b3o2bo$ob3obo3bobob3obo3bobo$o3bobo
3bobo3bobo3bobo$o2b2obo3bobo2b2obo3bobo$b2obo2b3o3b2obo2b3o2b4o!`)),
	}
}

var Library = &patternsLib{
	content: defaultPatternsLib(),
}

type patternsLib struct {
	content map[string]Pattern
	mutex   sync.RWMutex
}

func (l *patternsLib) Register(pattern Pattern) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.content[pattern.Name] = pattern
}

func (l *patternsLib) Get(name string) (Pattern, bool) {
	l.mutex.RLock()
	p, ok := l.content[name]
	l.mutex.RUnlock()
	return p, ok
}

func (l *patternsLib) All() map[string]Pattern {
	l.mutex.RLock()
	all := maps.Clone(l.content)
	l.mutex.RUnlock()
	return all
}

func (l *patternsLib) Filtered(fns []func(pattern Pattern) bool) map[string]Pattern {
	l.mutex.RLock()
	all := make(map[string]Pattern, len(l.content))
	for k, v := range l.content {
		ok := true
		for _, fn := range fns {
			if fn != nil && !fn(v) {
				ok = false
				break
			}
		}
		if ok {
			all[k] = v
		}
	}
	l.mutex.RUnlock()
	return all
}

func (l *patternsLib) Len() int {
	l.mutex.RLock()
	cl := len(l.content)
	l.mutex.RUnlock()
	return cl
}
