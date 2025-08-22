package db

import (
	"fmt"

	"github.com/NikolayProkopchuk/social/internal/store"
)

var usernames = []string{"Nikolay", "Ihor", "Sergey", "Vladimir", "Alexander", "Vlad"}

func Seed() error {
	generateUsers(10)
	return nil
}

func generateUsers(usersNum int) []*store.User {
	users := make([]*store.User, usersNum)

	for i := 0; i < usersNum; i++ {
		user := &store.User{
			Username: usernames[i%len(usernames)] + fmt.Sprintf("%d", i),
			Email:    usernames[i%len(usernames)] + fmt.Sprintf("%d", i) + "@test.com",
		}
		users[i] = user
	}

	return users
}
