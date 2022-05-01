package frequency_stat

import (
	"sync"
	"time"

	"github.com/saitofun/qlib/container/qlist"
	"github.com/saitofun/qlib/container/qtype"
)

type KeyPubTimeElement struct {
	Word string
	time time.Time
}

func NewKeyPubTimeElement(word string) *KeyPubTimeElement {
	return &KeyPubTimeElement{Word: word, time: time.Now()}
}

type KeyCountElement struct {
	Word  string
	count *qtype.Int
}

func NewKeyCountElement(word string) *KeyCountElement {
	return &KeyCountElement{Word: word, count: qtype.NewIntWithVal(1)}
}

func (e *KeyCountElement) Count() int { return e.count.Val() }

func (e *KeyCountElement) Add(delta int) { e.count.Add(delta) }

type zset struct {
	set      map[string]*qlist.Element
	sequence *qlist.List   // sequence 发布按序
	ordered  *qlist.List   // ordered 频次按序
	keep     time.Duration // keep 统计保存时长(sequence)
	mtx      *sync.Mutex
}

func NewSet(keep time.Duration) *zset {
	return &zset{
		set:      make(map[string]*qlist.Element),
		sequence: qlist.New(),
		ordered:  qlist.New(),
		keep:     keep,
		mtx:      &sync.Mutex{},
	}
}

// AddWords 向集合中添加单词
func (z *zset) AddWords(words ...string) {
	z.mtx.Lock()
	defer z.mtx.Unlock()

	for _, word := range words {
		z.sequence.PushBack(NewKeyPubTimeElement(word))
		if elem, ok := z.set[word]; ok {
			elem.Value.(*KeyCountElement).Add(1)
			z.reorderAfterCountIncr(elem)
		} else {
			added := NewKeyCountElement(word)
			z.set[word] = z.ordered.PushBack(added)
		}
	}
}

// check 检查过期单词
func (z *zset) check() {
	z.sequence.ReverseRange(func(elem *qlist.Element) bool {
		if elem == nil {
			return false
		}
		if v := elem.Value.(*KeyPubTimeElement); time.Since(v.time) > z.keep {
			pos := z.set[v.Word]
			val := pos.Value.(*KeyCountElement)
			val.Add(-1)
			if val.Count() == 0 {
				z.ordered.Remove(z.set[val.Word])
				delete(z.set, val.Word)
			}
			z.reorderAfterCountDecr(pos)
			return true
		}
		return false
	})
}

// ReorderAfterCountIncr 数量降低 重新调整排序
func (z *zset) reorderAfterCountIncr(elem *qlist.Element) {
	prev, count := elem.Prev(), elem.Value.(*KeyCountElement).Count()
	for ; prev != nil; prev = prev.Prev() {
		if count < prev.Value.(*KeyCountElement).Count() {
			break
		}
	}
	if prev == nil {
		z.ordered.MoveToFront(elem)
	} else {
		z.ordered.MoveAfter(elem, prev)
	}
}

// ReorderAfterCountDecr 数量增加 重新调整排序
func (z *zset) reorderAfterCountDecr(elem *qlist.Element) {
	next, count := elem.Next(), elem.Value.(*KeyCountElement).Count()
	for ; next != nil; next = next.Next() {
		if count > next.Value.(*KeyCountElement).Count() {
			break
		}
	}
	if next == nil {
		z.ordered.MoveToBack(elem)
	} else {
		z.ordered.MoveBefore(elem, next)
	}
}

// TopN 出现频次最多的N个单词
func (z *zset) TopN(n int) (ret []KeyCountElement) {
	z.mtx.Lock()
	defer z.mtx.Unlock()

	z.check()
	elements := z.ordered.FrontN(n)

	idx := 0
	for _, e := range elements {
		if idx == n {
			break
		}
		val := e.(*KeyCountElement)
		ret = append(ret, KeyCountElement{
			Word:  val.Word,
			count: val.count.Clone(),
		})
		idx++
	}
	return
}

func (z *zset) Top1() KeyCountElement { return z.TopN(1)[0] }
