package service

import (
	"fmt"
	"sync"
)

type UserStore interface {
	Save(user *User) error
	Find(username string) *User
}

type InMemoryUserStore struct {
	mutex sync.RWMutex
	users map[string]*User
}

func NewInMemoryUserStore() *InMemoryUserStore {
	return &InMemoryUserStore{
		users: make(map[string]*User),
	}
}

func (userStore *InMemoryUserStore) Save(user *User) error {
	userStore.mutex.Lock()
	defer userStore.mutex.Unlock()

	if user := userStore.users[user.Username]; user != nil {
		return ErrAlreadyExists
	}

	clone, err := user.Clone()
	if err != nil {
		return fmt.Errorf("cannot clone the user: %v", err)
	}

	userStore.users[user.Username] = clone
	return nil
}

func (userStore *InMemoryUserStore) Find(username string) *User {
	userStore.mutex.RLock()
	defer userStore.mutex.RUnlock()

	return userStore.users[username]
}
