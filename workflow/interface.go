package workflow

type Workable interface {
	Start() error
	Cancel() error
	Join() error
}

