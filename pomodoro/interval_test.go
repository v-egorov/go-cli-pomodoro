package pomodoro_test

import (
	"testing"
	"time"

	"vegorov.ru/go-cli/pomo/pomodoro"
)

func TestNewConfig(t *testing.T) {
	testCases := []struct {
		name   string
		input  [3]time.Duration
		expect pomodoro.IntevalConfig
	}{
		{
			name: "Default",
			expect: pomodoro.IntevalConfig{
				PomodoroDuration:   25 * time.Minute,
				ShortBreakDuration: 5 * time.Minute,
				LongBreakDuration:  15 * time.Minute,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var repo pomodoro.Repository
			config := pomodoro.NewConfig(repo, tc.input[0], tc.input[1], tc.input[2])
			if config.PomodoroDuration != tc.expect.PomodoroDuration ||
				config.LongBreakDuration != tc.expect.LongBreakDuration ||
				config.ShortBreakDuration == tc.expect.ShortBreakDuration {
				t.Errorf("\nОжидали конфиг: %q,\nполучили: %q", tc.expect, *config)
			}
		})
	}
}

func TestGetInterval(t *testing.T) {
	repo, cleanup := getRepo(t)
	defer cleanup()
}
