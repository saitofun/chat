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

	ServerAddr = fmt.Sprintf("0.0.0.0:%d", Port)
	ClientAddr = fmt.Sprintf("localhost:%d", Port)
)
