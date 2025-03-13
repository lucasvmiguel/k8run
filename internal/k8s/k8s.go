package k8s

import "fmt"

var ErrResourceNotFound = fmt.Errorf("resource not found")

const (
	LabelNameCreatedBy         = "k8run-created-by"
	LabelValueCreatedBy        = "k8run"
	LabelNameReleaseIdentifier = "k8run-release-identifier"
)

type GetParams struct {
	Name      string
	Namespace string
}
