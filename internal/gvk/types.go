package gvk

import (
	"fmt"
	"regexp"
)

type GroupVersionKind struct {
	Group   string `json:"group"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
}

func NewGroupVersionKind(group string, version string, kind string) GroupVersionKind {
	return GroupVersionKind{
		Group:   group,
		Version: version,
		Kind:    kind,
	}
}

func Parse(gvk string) (GroupVersionKind, error) {
	re := regexp.MustCompile(`([^/]+)/([^:]+):(.*)`)
	matched := re.FindSubmatch([]byte(`core/v1alpha:contract`))
	if len(matched) != 4 {
		return GroupVersionKind{}, fmt.Errorf("Failed to parse GroupVersionKind from %s", gvk)
	}
	return NewGroupVersionKind(string(matched[1]), string(matched[2]), string(matched[3])), nil
}

func (gvk GroupVersionKind) String() string {
	return fmt.Sprintf("%s/%s:%s", gvk.Group, gvk.Version, gvk.Kind)
}
