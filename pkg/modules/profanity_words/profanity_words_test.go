package profanity_words_test

import (
	"testing"

	. "github.com/saitofun/chat/pkg/modules/profanity_words"
	"github.com/stretchr/testify/require"
)

func TestMaskBy(t *testing.T) {
	var (
		dict   = []string{"c++", "c#", "java", "python", "just", "and", ",", "java and python"}
		input  = "/Generally, I just use golang, c++, c#, java and python for developing/"
		expect = "/Generally  I      use golang                           for developing/"
	)

	AddWords(dict...)

	tt := require.New(t)
	tt.Equal(expect, MaskWordsBy(input, ' '))

	AddWords("fuck")
	input = "/fuckyou/"
	expect = "/****you/"
	tt.Equal(expect, MaskWordsBy(input, '*'))
}
