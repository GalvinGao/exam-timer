package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/url"
	"os"
	"path"
)

type Session struct {
	Started  bool
	Timers   []Question
	Current  uint
	fileName string
}

func NewSession(count uint, fileName string) *Session {
	var questions []Question
	for i := uint(0); i < count; i++ {
		questions = append(questions, NewQuestion(i))
	}

	recordsFile, err := url.Parse("records")
	if err != nil {
		log.Fatal(err)
	}
	recordsFile.Path = path.Join(recordsFile.Path, fileName+".json")

	return &Session{
		Current:  0,
		Timers:   questions,
		fileName: recordsFile.String(),
	}
}

func (s *Session) NewTimerAt(index uint) {
	s.Timers[index] = NewQuestion(index + 1)
}

func (s *Session) CurrentTimer() *Question {
	return &(s.Timers[s.Current])
}

func (s *Session) Start() error {
	if !s.Started {
		s.NewTimerAt(0)
		err := s.CurrentTimer().Start()
		if err != nil {
			return err
		}
	} else {
		return errors.New("session has already started")
	}
	s.Started = true
	return nil
}

func (s *Session) Next() error {
	err := s.CurrentTimer().Stop()
	if err != nil {
		return err
	}
	s.Current++
	s.NewTimerAt(s.Current)
	err = s.CurrentTimer().Start()
	if err != nil {
		return err
	}
	return nil
}

func (s *Session) End() error {
	_ = s.CurrentTimer().Stop()
	file, err := os.Create(s.fileName)
	if err != nil {
		return err
	}
	set := map[uint]int64{}
	for i := range s.Timers {
		timer := s.Timers[i]
		set[timer.Index] = -timer.Duration.Nanoseconds()
	}
	jsonBytes, err := json.Marshal(set)
	if err != nil {
		return err
	}
	_, err = file.Write(jsonBytes)
	if err != nil {
		return err
	}
	err = file.Close()
	return err
}

func (s *Session) SwitchPause() error {
	if s.CurrentTimer().Running() {
		err := s.CurrentTimer().Stop()
		s.Started = false
		if err != nil {
			return err
		}
	} else {
		err := s.CurrentTimer().Start()
		s.Started = true
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Session) GetEdited() uint {
	counter := uint(0)
	for i := range s.Timers {
		if s.Timers[i].Edited {
			counter++
		}
	}
	return counter
}
