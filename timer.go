package main

import (
	"errors"
	"time"
)

const (
	QuestionTimerStarted = iota
	QuestionTimerStopped
)

type Question struct {
	Status    uint          `json:"status"`
	Index     uint          `json:"index"`
	Edited    bool          `json:"edited"`
	Duration  time.Duration `json:"duration"`
	LastStart time.Time
}

func NewQuestion(index uint) Question {
	return Question{
		Status:    QuestionTimerStopped,
		Index:     index,
		LastStart: time.Unix(0, 0),
	}
}

func (q *Question) Start() error {
	if q.Status == QuestionTimerStarted {
		return errors.New("timer has already started")
	}
	q.LastStart = time.Now()
	q.Status = QuestionTimerStarted
	q.Edited = true
	return nil
}

func (q *Question) Stop() error {
	if q.Status == QuestionTimerStopped {
		return errors.New("timer has already stopped")
	}
	now := time.Now()
	q.Duration = now.Add(q.Duration).Add(q.LastStart.Sub(now)).Sub(now)
	q.Status = QuestionTimerStopped
	return nil
}

func (q *Question) Running() bool {
	return q.Status == QuestionTimerStarted
}
