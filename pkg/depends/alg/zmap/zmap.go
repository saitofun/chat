// Package zmap Impls an ordered map access
package zmap

import (
	"sort"
	"strings"
)

type Map struct {
	keys []interface{}
	vals map[interface{}]interface{}
}

var _ sort.Interface = (*Map)(nil)

func (m *Map) Len() int           { return 0 }
func (m *Map) Less(i, j int) bool { return false }
func (m *Map) Swap(i, j int)      {}

func OnPub(sentences string) {
	words := strings.Split(sentences, " ")
	for i := range words {
		words[i] = strings.TrimSpace(words[i])
	}
}
