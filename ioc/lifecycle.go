package ioc

type Startable interface {
	StartComponent() error
}

type Stoppable interface {
	PrepareToStop()
	ReadyToStop() (bool, error)
	Stop() error
}

type AccessibilityBlocker interface {
	BlockAccess() (bool, error)
}

type Accessible interface {
	AllowAccess() error
}
