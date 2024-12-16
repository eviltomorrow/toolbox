package rule

import "time"

type Rule struct {
	Name     string
	Port     int
	Command  string
	Duration time.Duration
}

var cache = make(map[string]*Rule)

func Register()
