// Package zmap Impls an ordered map access
package zmap

import (
	"sort"
)

type Map struct {
	keys []interface{}
	vals map[interface{}]interface{}
}

var _ sort.Interface = (*Map)(nil)

func (m *Map) Len() int           { return 0 }
func (m *Map) Less(i, j int) bool { return false }
func (m *Map) Swap(i, j int)      {}
