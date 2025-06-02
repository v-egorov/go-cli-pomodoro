package pomodoro_test

import (
	"testing"

	"vegorov.ru/go-cli/pomo/pomodoro"
	"vegorov.ru/go-cli/pomo/pomodoro/repository"
)

// Helper function - возвращает репозиторий для тестов
func getRepo(t *testing.T) (pomodoro.Repository, func()) {
	t.Helper()
	// для in-memory не требуется cleanup function, поэтому возвращаем пустую func() {}
	return repository.NewInMemoryRepo(), func() {}
}
