package main

import (
	"sort"
)

type SortedMap struct {
	Keys []string
	Vals []int
}

func NewSortedMap(m map[string]int) *SortedMap {
	sm := &SortedMap{
		Keys: make([]string, 0, len(m)),
		Vals: make([]int, 0, len(m)),
	}
	for k, v := range m {
		sm.Keys = append(sm.Keys, k)
		sm.Vals = append(sm.Vals, v)
	}
	return sm
}

func (sm *SortedMap) Sort() {
	sort.Sort(sm)
}

func (sm *SortedMap) Len() int {
	return len(sm.Vals)
}

func (sm *SortedMap) Less(i, j int) bool {
	return sm.Vals[i] < sm.Vals[j]
}

func (sm *SortedMap) Swap(i, j int) {
	sm.Vals[i], sm.Vals[j] = sm.Vals[j], sm.Vals[i]
	sm.Keys[i], sm.Keys[j] = sm.Keys[j], sm.Keys[i]
}
