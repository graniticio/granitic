package ioc

type LifecycleSupport int

const (
	CanStart = iota
	CanStop
	CanSuspend
	CanBlockStart
	CanBeAccessed
)

type Startable interface {
	StartComponent() error
}

type Suspendable interface {
	Suspend() error
	Resume() error
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
