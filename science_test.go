package science

import (
	"testing"
)

func TestExperimentChecksFunctions(t *testing.T) {
	e := NewExperiment("test")

	err := e.Run()
	if err != ErrNoControl {
		t.Fatal("expected control to be required")
	}

	e.Control = func() interface{} { return nil }
	err = e.Run()
	if err != ErrNoCandidate {
		t.Fatal("expected candidate to be required")
	}

	e.Candidate = func() interface{} { return nil }
	e.Comparator = nil
	err = e.Run()
	if err != ErrNoComparator {
		t.Fatal("expected comparator to be required")
	}
}

func TestExperimentRunsControlIfNotEnabled(t *testing.T) {
	var ran = false

	e := NewExperiment("test")
	e.Control = func() interface{} {
		ran = true
		return nil
	}
	e.Candidate = func() interface{} { return nil }

	e.Run()

	if !ran {
		t.Fatal("expected control to run")
	}
}

func TestExperimentRunsControlAndCandidate(t *testing.T) {
	var controlRan = false
	var candidateRan = false

	e := NewExperiment("test")
	e.Control = func() interface{} {
		controlRan = true
		return nil
	}
	e.Candidate = func() interface{} {
		candidateRan = true
		return nil
	}

	e.Run()

	if !controlRan {
		t.Fatal("expected control to run")
	}

	if !candidateRan {
		t.Fatal("expected candidate to run")
	}
}

func TestExperimentPublishesSuccess(t *testing.T) {
	e := NewExperiment("test experiment")

	e.Control = func() interface{} {
		return 42
	}
	e.Candidate = func() interface{} {
		return 42
	}

	var result *Result
	e.Publish = func(p *Result) {
		result = p
	}

	e.Run()

	if result.Name != "test experiment" {
		t.Fatal("expected result to contain the experiment name")
	}

	if !result.Matched {
		t.Fatal("expected published results to be a match")
	}

	if result.Control.Value.(int) != 42 {
		t.Fatal("expected result control result to contain the value")
	}

	if result.Candidate.Value.(int) != 42 {
		t.Fatal("expected result candidate result to contain the value")
	}
}

func TestExperimentPublishesFailure(t *testing.T) {
	e := NewExperiment("name")

	e.Control = func() interface{} {
		return 42
	}
	e.Candidate = func() interface{} {
		return 43
	}

	var matched bool
	e.Publish = func(result *Result) {
		matched = result.Matched
	}

	e.Run()

	if matched {
		t.Fatal("expected published results to be a mismatch")
	}
}
