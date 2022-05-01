package util_test

import (
	"testing"

	"github.com/saitofun/chat/cmd/config"
	. "github.com/saitofun/chat/pkg/depends/util"
)

func TestDownloadFile(t *testing.T) {
	if err := DownloadFile(
		config.RemoteProfanityWordsURL,
		"./example/profanity_words.txt",
	); err != nil {
		t.Log(err)
	}
}
