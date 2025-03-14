package k8s

import "fmt"

// ErrResourceNotFound is the error returned when a resource is not found.
var ErrResourceNotFound = fmt.Errorf("resource not found")

const (
	// LabelNameCreatedBy is the label name to identify resources created by k8run.
	LabelNameCreatedBy = "k8run-created-by"
	// LabelValueCreatedBy is the label value to identify resources created by k8run.
	LabelValueCreatedBy = "k8run"
	// LabelNameReleaseIdentifier is the label name to identify resources by release identifier.
	LabelNameReleaseIdentifier = "k8run-release-identifier"
)

// GetParams represents the parameters to get a resource.
type GetParams struct {
	Name      string
	Namespace string
}
