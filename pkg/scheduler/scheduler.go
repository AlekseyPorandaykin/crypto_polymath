package scheduler

import (
	"context"
	"time"
)

func ExecuteCustomMinuteWithReply(ctx context.Context, minute, slippageSec, intervalSecond, maxIteration int, fn func() (bool, error)) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			ticker.Stop()
			i := 0
			for {
				ok, err := fn()
				if err != nil {
					return err
				}
				i++
				if ok {
					break
				}
				if i >= maxIteration {
					break
				}
				time.Sleep(time.Duration(intervalSecond) * time.Second)
			}
			ticker.Reset(durationToMinute(time.Now(), minute) + time.Duration(slippageSec)*time.Second)
		}
	}
}

func ExecuteCustomMinute(ctx context.Context, minute, addSecond int, fn func() error) error {
	_ = fn()
	ticker := time.NewTicker(durationToMinute(time.Now(), minute) + time.Duration(addSecond)*time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			ticker.Stop()
			if err := fn(); err != nil {
				return err
			}
			ticker.Reset(durationToMinute(time.Now(), minute) + time.Duration(addSecond)*time.Second)
		}
	}
}

func ExecuteEveryHour(ctx context.Context, hours, addSecond int, fn func() error) error {
	_ = fn()
	slippageSecond := time.Duration(addSecond) * time.Second
	ticker := time.NewTicker(durationToHour(time.Now(), hours) + slippageSecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			ticker.Stop()
			if err := fn(); err != nil {
				return err
			}
			ticker.Reset(durationToHour(time.Now(), hours) + slippageSecond)
		}
	}
}

func ExecuteEveryDay(ctx context.Context, fn func() error) error {
	_ = fn()
	ticker := time.NewTicker(durationToNextDay(time.Now()))
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			ticker.Stop()
			if err := fn(); err != nil {
				return err
			}
			ticker.Reset(durationToNextDay(time.Now()))
		}
	}
}

func ExecuteEveryWeek(ctx context.Context, fn func() error) error {
	_ = fn()
	ticker := time.NewTicker(durationToNextWeek(time.Now()))
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			ticker.Stop()
			if err := fn(); err != nil {
				return err
			}
			ticker.Reset(durationToNextWeek(time.Now()))
		}
	}
}

func ExecuteEveryMonth(ctx context.Context, fn func() error) error {
	_ = fn()
	ticker := time.NewTicker(durationToNextWeek(time.Now()))
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			ticker.Stop()
			if err := fn(); err != nil {
				return err
			}
			ticker.Reset(durationToNextWeek(time.Now()))
		}
	}
}

func durationToMinute(from time.Time, minute int) time.Duration {
	now := from.In(time.UTC)
	if minute >= 60 || minute <= 0 {
		return time.Minute
	}
	nextExecute := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, minute, 0, 0, time.UTC)
	for i := 0; i <= 60; i += minute {
		if now.Minute() < i {
			nextExecute = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), i, 0, 0, time.UTC)
			break
		}
	}
	return nextExecute.Sub(now)
}

func durationToHour(from time.Time, hour int) time.Duration {
	now := from.In(time.UTC)
	if hour <= 0 || hour >= 24 {
		return time.Minute
	}
	nextExecute := time.Date(now.Year(), now.Month(), now.Day()+1, 5, 0, 0, 0, time.UTC)
	for i := 0; i <= 24; i += hour {
		if now.Hour() < i {
			nextExecute = time.Date(now.Year(), now.Month(), now.Day(), i, 0, 0, 0, time.UTC)
			break
		}
	}
	return nextExecute.Sub(now)
}

func durationToNextHour(from time.Time) time.Duration {
	now := from.In(time.UTC)
	nextExecute := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 0, time.UTC)
	return nextExecute.Sub(now)
}
func durationToNextDay(from time.Time) time.Duration {
	now := from.In(time.UTC)
	nextExecute := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
	return nextExecute.Sub(now)
}
func durationToNextWeek(from time.Time) time.Duration {
	now := from.In(time.UTC)
	day := 7 - int(now.Weekday())
	nextExecute := time.Date(now.Year(), now.Month(), now.Day()+day+1, 0, 0, 0, 0, time.UTC)
	return nextExecute.Sub(now)
}
func durationToNextMonth(from time.Time) time.Duration {
	now := from.In(time.UTC)
	nextExecute := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	return nextExecute.Sub(now)
}
