package repository

import (
	"sync"

	"vegorov.ru/go-cli/pomo/pomodoro"
)

// Репозиторий для работы с интервалами в памяти
type inMemoryRepo struct {
	sync.RWMutex
	inetravls []pomodoro.Interval
}
