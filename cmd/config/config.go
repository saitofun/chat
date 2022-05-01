package config

import (
	"fmt"
	"time"
)

var (
	MaxRoomCache             = 50
	MaxRoomPopularWords      = 10
	PopularWordsKeepDuration = time.Minute * 10
	Addr                     = "localhost"
	Port                     = 10086

	ServerAddr = fmt.Sprintf(":%d", Port)
	ClientAddr = fmt.Sprintf("localhost:%d", Port)

	RemoteProfanityWordsURL = "https://raw.githubusercontent.com/CloudcadeSF/google-profanity-words/main/data/list.txt"
	LocalProfanityWordsPath = "config/profanity_words.txt"
	ProfanityWordsMask      = rune('*')
)
