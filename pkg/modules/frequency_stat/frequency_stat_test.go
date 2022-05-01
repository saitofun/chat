package frequency_stat

import (
	"testing"
	"time"

	"github.com/saitofun/qlib/container/qlist"
	"github.com/stretchr/testify/require"
)

var (
	keep = time.Minute
	z    = NewSet(time.Minute * 10)
)

func TestZset_AddWords(t *testing.T) {
	tt := require.New(t)
	z.AddWords("great", "fine", "fine", "good", "good", "good")

	{
		expectedWord := []string{"good", "fine", "great"}
		expectedCount := []int{3, 2, 1}
		expectedLen := 0
		idx := 0

		z.ordered.Range(func(e *qlist.Element) bool {
			if e == nil {
				tt.Equal(expectedLen, idx)
				return false
			}
			v := e.Value.(*KeyCountElement)
			tt.Equal(expectedWord[idx], v.Word)
			tt.Equal(expectedCount[idx], v.Count())
			idx++
			return true
		})
	}

	top1 := z.TopN(1)

	tt.Equal(top1[0].Word, "good")
	tt.Equal(top1[0].Count(), 3)
}
