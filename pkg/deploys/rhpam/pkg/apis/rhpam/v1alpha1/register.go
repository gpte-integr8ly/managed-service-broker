package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	group   = "rhpam.integreatly.org"
	version = "v1alpha1"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion  = schema.GroupVersion{Group: group, Version: version}
	SchemeGroupResource = schema.GroupResource{Group: group, Resource: NewRhpamDev().Kind}
)

func ApiVersion() string {
	return group + "/" + version
}
