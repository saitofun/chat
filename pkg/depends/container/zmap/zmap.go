// Package zmap Impls a ordered map access
package zmap

import "github.com/saitofun/qlib/container/qlist"

type OrderedMap struct {
	val map[interface{}]interface{}
	lst map[string]*qlist.List
}

func New() {
}
