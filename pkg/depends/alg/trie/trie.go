package trie

type Node struct {
	End  bool
	Leaf map[rune]*Node
}

func NewNode() *Node {
	return &Node{false, make(map[rune]*Node)}
}

// Root with a node filed as a super root
type Root struct{ *Node }

func NewRoot(words ...string) *Root {
	root := &Root{NewNode()}
	root.Build(words)
	return root
}

func (r *Root) Build(words []string) {
	for _, word := range words {
		r.Insert([]rune(word))
	}
}

func (r *Root) Insert(word []rune) {
	iter := r.Node
	for _, b := range word {
		if next, ok := iter.Leaf[b]; !ok {
			next = NewNode()
			iter.Leaf[b] = next
			iter = next
		} else {
			iter = next
		}
	}
	iter.End = true
}

func (r *Root) StartWith(prefix []rune) bool {
	iter := r.Node
	for _, b := range prefix {
		if next, ok := iter.Leaf[b]; !ok {
			return false
		} else {
			iter = next
		}
	}
	return true
}

func (r *Root) SearchWord(word []rune) bool {
	iter := r.Node
	for _, b := range word {
		if next, ok := iter.Leaf[b]; !ok {
			return false
		} else {
			iter = next
		}
	}
	return iter.End
}

func (r *Root) ScanSentence(s string) (matched []Matcher) {
	sentence := []rune(s)
	start, end, found := 0, 1, false

	for end <= len(sentence) {
		if r.StartWith(sentence[start:end]) {
			found = true
			end++
			continue
		}
		if found {
			for idx := end - 1; idx > start; idx-- {
				if r.SearchWord(sentence[start:idx]) {
					matched = append(matched, Matcher{start, idx})
					start = idx
					end = idx + 1
					break
				}
			}
		} else {
			start++
			end = start + 1
		}
		found = false
	}

	if found {
		for idx := end - 1; idx > start; idx-- {
			if r.SearchWord(sentence[start:idx]) {
				matched = append(matched, Matcher{start, idx - 1})
			}
		}
	}
	return
}

// Matcher sub slice range marker, which identified as [Match[0]:Matcher[1]]
type Matcher [2]int

func (m Matcher) IsZero() bool { return m[0] == 0 && m[1] == 0 }

func (m Matcher) IsValid() bool { return m[0] >= 0 && m[1] >= 0 && m[0] <= m[1] }

func (m Matcher) Sub(s []any) []any { return s[m[0] : m[1]+1] }
