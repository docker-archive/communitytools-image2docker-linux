package workflow

import "errors"

type defaultOrchestrator struct {}

var errNotYetImplemented = errors.New(`not yet implemented`)

func (o *defaultOrchestrator) Start() error { return errNotYetImplemented }
func (o *defaultOrchestrator) Cancel() error { return errNotYetImplemented }
func (o *defaultOrchestrator) Join() error { return errNotYetImplemented }

var d defaultOrchestrator = defaultOrchestrator{}

func Start() error {
	return d.Start()
}

func Cancel() error {
	return d.Cancel()
}

func NewWorkflow() Workable {
	return &defaultOrchestrator{}
}
