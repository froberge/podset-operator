package v1alpha1

// ContainerWaitingReason is a label for the reason a container is in waiting state
type ContainerWaitingReason = string

// These are the valid reason for the container waiting
const (
	ContainerCreating          ContainerWaitingReason = "ContainerCreating"
	CrashLoopBackOff           ContainerWaitingReason = "CrashLoopBackOff"
	ErrImagePull               ContainerWaitingReason = "ErrImagePull"
	ImagePullBackOff           ContainerWaitingReason = "ImagePullBackOff"
	CreateContainerConfigError ContainerWaitingReason = "CreateContainerConfigError"
	InvalidImageName           ContainerWaitingReason = "InvalidImageName"
	CreateContainerError       ContainerWaitingReason = "CreateContainerError"
)

// ContainerTerminatedReason is a label for the reason a container is in terminated state
type ContainerTerminatedReason = string

// These are the valid reason for the container terminated
const (
	Terminated         ContainerTerminatedReason = "Terminated"
	OOMKilled          ContainerTerminatedReason = "OOMKilled"
	Error              ContainerTerminatedReason = "Error"
	Completed          ContainerTerminatedReason = "Completed"
	ContainerCannotRun ContainerTerminatedReason = "ContainerCannotRun"
	DeadlineExceeded   ContainerTerminatedReason = "DeadlineExceeded"
)
