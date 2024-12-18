package user

import "github.com/eviltomorrow/toolbox/apps/ssh-server/conf"

var cache = map[string]string{}

func Auth(username, password string) bool {
	value, ok := cache[username]
	if !ok {
		return false
	}
	if value == password {
		return true
	}
	return false
}

func Load(users map[string]conf.User) {
	for _, user := range users {
		cache[user.Username] = user.Password
	}
}
