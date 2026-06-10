package runtime

// Status is the process status class shared by scip-search runtime paths.
type Status int

const (
	StatusOK         Status = 0
	StatusUsage      Status = 2
	StatusIndexLoad  Status = 3
	StatusValidation Status = 4
)

// Failure identifies a shared runtime failure without defining command payloads.
type Failure struct {
	status  Status
	message string
}

// UsageFailure classifies an invocation-shape failure.
func UsageFailure(message string) Failure {
	return Failure{
		status:  StatusUsage,
		message: message,
	}
}

// IndexLoadFailure classifies a failure while loading caller-selected SCIP input.
func IndexLoadFailure(message string) Failure {
	return Failure{
		status:  StatusIndexLoad,
		message: message,
	}
}

// ValidationFailure classifies valid input that fails command-specific validation.
func ValidationFailure(message string) Failure {
	return Failure{
		status:  StatusValidation,
		message: message,
	}
}

// Status returns the documented process status for the failure class.
func (failure Failure) Status() Status {
	return failure.status
}

// Message returns the human-readable diagnostic for stderr.
func (failure Failure) Message() string {
	return failure.message
}

// Error lets shared runtime failures cross Go error-returning boundaries.
func (failure Failure) Error() string {
	return failure.message
}
