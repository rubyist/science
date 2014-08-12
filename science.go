package science

import (
	"errors"
	"math/rand"
	"reflect"
	"time"
)

type Experiment struct {
	Name         string
	Control      func() interface{}
	Candidate    func() interface{}
	Comparator   func(interface{}, interface{}) bool
	Enabled      func() bool
	Publish      func(*Payload)
	controlFirst bool
}

type Payload struct {
	Name         string
	Timestamp    time.Time
	ControlFirst bool
	Matched      bool
	Candidate    *Result
	Control      *Result
}

type Result struct {
	Duration time.Duration
	Value    interface{}
}

var (
	ErrNoControl    = errors.New("control function missing")
	ErrNoCandidate  = errors.New("candidate function missing")
	ErrNoComparator = errors.New("comparator function missing")
)

func enabledByDefault() bool { return true }

func NewExperiment(name string) *Experiment {
	controlFirst := rand.Intn(2) == 0
	return &Experiment{
		Name:         name,
		Comparator:   reflect.DeepEqual,
		controlFirst: controlFirst,
		Enabled:      enabledByDefault}
}

func (e *Experiment) Run() error {
	if e.Control == nil {
		return ErrNoControl
	}
	if e.Candidate == nil {
		return ErrNoCandidate
	}
	if e.Comparator == nil {
		return ErrNoComparator
	}

	if e.Enabled == nil || !e.Enabled() {
		e.Control()
		return nil
	}

	ts := time.Now()
	var control *Result
	var candidate *Result

	// Should swallow any panics by Candidate
	if e.controlRunsFirst() {
		control = observe(e.Control)
		candidate = observe(e.Candidate)
	} else {
		candidate = observe(e.Candidate)
		control = observe(e.Control)
	}

	matched := e.Comparator(control.Value, candidate.Value)

	if e.Publish != nil {
		payload := &Payload{
			Name:         e.Name,
			Matched:      matched,
			ControlFirst: e.controlRunsFirst(),
			Timestamp:    ts,
			Candidate:    candidate,
			Control:      control,
		}
		e.Publish(payload)
	}

	return nil
}

func (e *Experiment) controlRunsFirst() bool {
	return true
}

func observe(f func() interface{}) *Result {
	start := time.Now()

	val := f()

	duration := time.Since(start)

	return &Result{
		Duration: duration,
		Value:    val}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
