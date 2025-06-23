package cmd

import (
	"vegorov.ru/go-cli/pomo/pomodoro"
	"vegorov.ru/go-cli/pomo/pomodoro/repository"
)

func getRepo() (pomodoro.Repository, error) {
	return repository.NewInMemoryRepo(), nil
}
