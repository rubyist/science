// package science is go package for measuring and validating code changes
// without altering behavior.
//
// Experiments begin with a piece of code that acts as a control, i.e. its
// behavior is generally known, and a piece of code that is the candidate, i.e.
// a refactoring of the control.
//
// When an Experiment runs, the following happens:
//  * Decision to run the candidate code
//  * Candidate code will run before control 50% of the time
//  * Duration of both behaviors are measured
//  * Return values of both behaviors are compared
//  * Results published via a Publish function
//
//
// A simple example:
//    func myFunc() {
//      e := science.NewExperiment("refactor.myFunc")
//      e.Control = func() interface{} {
//
//      }
//
//      e.Candidate = func() interface {
//
//      }
//
//      e.Publish = func(*science.Result) {
//
//      }
//
//      err := e.Run()
//
//    }
package science

import (
	"errors"
	"math/rand"
	"reflect"
	"time"
)

// Errors returned by Run
var (
	ErrNoControl    = errors.New("control function missing")
	ErrNoCandidate  = errors.New("candidate function missing")
	ErrNoComparator = errors.New("comparator function missing")
)

// The ExperimentFunc type is a function containing the code for the control
// and candidate of the experiment. This function can return any value. The
// values returned from the control and candidate functions are compared using
// a ComparatorFunc.
type ExperimentFunc func() interface{}

// The ComparatorFunc type is a function which compares the return values of
// the Control and Candidate functions. By default, reflect.DeepEqual is used.
type ComparatorFunc func(interface{}, interface{}) bool

// The EnabledFunc type is a function which  determines if the expermint is to
// be run. By default, this is a function that always returns true. If the
// function is nil or returns false, the Control will be run without any
// observation and the Candidate will not be run. If it returns true, both will
// be run.
type EnabledFunc func() bool

// PublishFunc is a function that receives the results Result of the experiment.
type PublishFunc func(*Result)

// Experiment is the experiment to run.
type Experiment struct {
	Name         string
	Control      ExperimentFunc
	Candidate    ExperimentFunc
	Comparator   ComparatorFunc
	Enabled      EnabledFunc
	Publish      PublishFunc
	controlFirst bool
}

// Result is the result sent to the Publish function, if one is provided.
type Result struct {
	Name         string       // Name of the experiment
	Timestamp    time.Time    // Time the experiment started
	ControlFirst bool         // Whether the Control ran before the Candidate
	Matched      bool         // Whether the control and candidate values matched
	Control      *Observation // Control results
	Candidate    *Observation // Candidate results
}

// Observation stores the results of running the Control or Candidate functions.
type Observation struct {
	Duration time.Duration // Duration of the function call
	Value    interface{}   // Return value of the function
}

// NewExperiment creates a new Experiment with the given name. The default
// Comparator function iw reflect.DeepEqual. The experiment is Enabled by
// default.
func NewExperiment(name string) *Experiment {
	controlFirst := rand.Intn(2) == 0
	return &Experiment{
		Name:         name,
		Comparator:   reflect.DeepEqual,
		Enabled:      enabledByDefault,
		controlFirst: controlFirst}
}

// Run runs the experiment. If any of the Control, Candidate, or Comparator are
// nil, Run will return an appropriate error.
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
	var control *Observation
	var candidate *Observation

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
		result := &Result{
			Name:         e.Name,
			Matched:      matched,
			ControlFirst: e.controlRunsFirst(),
			Timestamp:    ts,
			Candidate:    candidate,
			Control:      control,
		}
		e.Publish(result)
	}

	return nil
}

func (e *Experiment) controlRunsFirst() bool {
	return true
}

func observe(f func() interface{}) *Observation {
	start := time.Now()

	val := f()

	duration := time.Since(start)

	return &Observation{
		Duration: duration,
		Value:    val}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func enabledByDefault() bool { return true }
