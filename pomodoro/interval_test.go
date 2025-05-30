package pomodoro_test

import (
	"context"
	"errors"
	"fmt"
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
				config.ShortBreakDuration != tc.expect.ShortBreakDuration {
				t.Errorf("\nОжидали конфиг: %q,\nполучили: %q", tc.expect, *config)
			}
		})
	}
}

func TestGetInterval(t *testing.T) {
	repo, cleanup := getRepo(t)
	defer cleanup()

	const duration = 1 * time.Millisecond
	config := pomodoro.NewConfig(repo, 3*duration, duration, 2*duration)

	for i := 1; i <= 16; i++ {
		// Чтобы покрыть тестом все сценарии получения интервалов -
		// нужно пробежать 2 полных цикла по 8 интервалов
		// Почему по 8 - см комментарии ниже в case i%8 == 0
		var (
			expCategory string
			expDuration time.Duration
		)

		switch {
		case i%2 != 0: // каждый нечётный интервал - это pomodoro
			expCategory = pomodoro.CategoryPomodoro
			expDuration = 3 * duration
		case i%8 == 0:
			// p - sb - p - sb - p - sb - p - lb
			// 1   2    3   4    5   6    7   8
			// каждый 8-й интервад - LongBreak
			expCategory = pomodoro.CategoryLongBreak
			expDuration = 2 * duration
		case i%2 == 0:
			// каждый чётный, но не каждый 8-й - ShortBreak
			expCategory = pomodoro.CategoryShortBreak
			expDuration = duration
		}

		testName := fmt.Sprintf("%s%d", expCategory, i)
		t.Run(testName, func(t *testing.T) {
			testInteval, err := pomodoro.GetInterval(config)
			if err != nil {
				t.Errorf("Не ожидали ошибку, а nполучили: %q", err)
			}

			emptyF := func(pomodoro.Interval) {}

			// При старте интервала он записывается в репозиторий,
			// поэтому дальше мы уже запрашиваем его по ID из репозитория,
			// чтобы проверить, что тестовый интевал завершился
			if err := testInteval.Start(context.Background(), config,
				emptyF, emptyF, emptyF); err != nil {
				t.Fatal(err)
			}

			if testInteval.Category != expCategory {
				t.Errorf("Ожидали категорию: %q, а получили: %q", expCategory, testInteval.Category)
			}

			if testInteval.PlannedDuration != expDuration {
				t.Errorf("Ожидали продолжительность интервала: %q, а получили: %q",
					expDuration, testInteval.PlannedDuration)
			}

			if testInteval.State != pomodoro.StateNotStarted {
				t.Errorf("Ожидали состояние интервала: %q, а получили: %q",
					pomodoro.StateNotStarted, testInteval.State)
			}

			// Тут как раз считываем из репозитория
			ui, err := repo.ByID(testInteval.ID)
			if err != nil {
				t.Errorf("Не ожидали ошибку, а получили: %q", err)
			}
			if ui.State != pomodoro.StateDone {
				t.Errorf("Ожидали состояние интервала: %q, а получили: %q",
					pomodoro.StateDone, ui.State)
			}
		})
	}
}

func TestPause(t *testing.T) {
	// Минимальная продолжительность интервала для этого теста - 2 секунды,
	// потому что tick() стартует тикер, который срабатывает раз в секунду.
	// Если продолжительность интервала будет менее 2 секунд, то мы не успеем
	// вызывть Pause(), а интервал уже завершится.
	const duration = 2 * time.Second
	// const duration = 1900 * time.Millisecond

	repo, cleanup := getRepo(t)
	defer cleanup()

	config := pomodoro.NewConfig(repo, duration, duration, duration)

	testCases := []struct {
		name        string
		start       bool
		expState    int
		expDuration time.Duration
	}{
		{
			name:        "NotStarted",
			start:       false,
			expState:    pomodoro.StateNotStarted,
			expDuration: 0,
		},
		{
			name:        "Paused",
			start:       true,
			expState:    pomodoro.StatePaused,
			expDuration: duration / 2,
		},
	}

	expError := pomodoro.ErrIntervalNotRunning

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			i, err := pomodoro.GetInterval(config)
			if err != nil {
				t.Fatal(err)
			}

			// 3 callbacks для запуска интервала
			start := func(pomodoro.Interval) {}
			end := func(pomodoro.Interval) {
				t.Errorf("End Callback не должен вызываться")
			}
			periodic := func(i pomodoro.Interval) {
				if err := i.Pause(config); err != nil {
					t.Fatal(err)
				}
			}

			if tc.start {
				if err := i.Start(ctx, config, start, periodic, end); err != nil {
					t.Fatal(err)
				}
			}

			i, err = pomodoro.GetInterval(config)
			if err != nil {
				t.Fatal(err)
			}

			err = i.Pause(config)
			if err != nil {
				if !errors.Is(err, expError) {
					t.Fatalf("Ожидали ошибку: %q, а получили: %q", expError, err)
				}
			}

			if err == nil {
				t.Errorf("Ожидали ошибку: %q, а получили nil", expError)
			}

			i, err = repo.ByID(i.ID)
			if err != nil {
				t.Fatal(err)
			}

			if i.State != tc.expState {
				t.Errorf("Ожидали состояние интервала: %d, а получили: %d", tc.expState, i.State)
			}

			if i.ActualDuration != tc.expDuration {
				t.Errorf("Ожидали продолжительность: %q, а получили: %q", tc.expDuration, i.ActualDuration)
			}

			cancel()
		})
	}
}
