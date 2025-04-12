package telegram

import (
	"sync"
)

type Users struct {
	mu    sync.RWMutex
	users map[int64]struct{}
}

func NewUsers() *Users {
	return &Users{
		users: make(map[int64]struct{}),
	}
}

func (u *Users) Add(userID int64) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.users[userID] = struct{}{}
}

func (u *Users) GetAll() []int64 {
	u.mu.RLock()
	defer u.mu.RUnlock()
	
	users := make([]int64, 0, len(u.users))
	for userID := range u.users {
		users = append(users, userID)
	}
	return users
} 