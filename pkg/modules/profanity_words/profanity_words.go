package profanity_words

import (
	"bufio"
	"io"
	"log"
	"os"

	"github.com/saitofun/chat/cmd/config"
	"github.com/saitofun/chat/pkg/depends/alg/trie"
	"github.com/saitofun/chat/pkg/depends/util"
)

var root *trie.Root

func init() {
	err := util.DownloadFile(
		config.RemoteProfanityWordsURL, config.LocalProfanityWordsPath,
	)
	if err != nil && !util.IsExist(config.LocalProfanityWordsPath) {
		log.Fatal("profanity words filtering disabled, because no dict file")
	}
	if err = LoadDictFromFile(config.LocalProfanityWordsPath); err != nil {
		log.Fatal("profanity words filtering disabled, because cannot load words")
	}
}

func LoadDictByWords(words []string) {
	root = trie.NewRoot(words...)
	for _, word := range words {
		root.Insert([]rune(word))
	}
}

func LoadDictFromFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	reader := bufio.NewReader(f)
	words := make([]string, 0)
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		words = append(words, string(line))
	}
	LoadDictByWords(words)
	return nil
}

func MaskWordsBy(src string, replacer rune) string {
	matched := root.ScanSentence(src)
	sentence := []rune(src)
	for _, pair := range matched {
		for idx := pair[0]; idx <= pair[1]; idx++ {
			sentence[idx] = replacer
		}
	}
	return string(sentence)
}

func MatchedWords(src string) []string {
	matched := root.ScanSentence(src)
	sentence := []rune(src)
	ret := make([]string, 0, len(matched))
	for _, pair := range matched {
		ret = append(ret, string(sentence[pair[0]:pair[1]]))
	}
	return ret
}

func AddWords(words ...string) {
	for _, word := range words {
		root.Insert([]rune(word))
	}
}
