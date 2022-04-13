package reaper

import (
	"errors"
	"flag"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

var ErrNotAVAlidTarget = errors.New("not a valid target")

var _ flag.Value = (*Target)(nil)

func ParseTarget(val string) (Target, error) {
	t := Target(val)
	if _, err := t.Selector(); err != nil {
		return "", err
	}
	return t, nil
}

type Target string

func (t Target) Selector() (labels.Selector, error) {
	s := string(t)
	if s == "" {
		return nil, ErrNotAVAlidTarget
	}

	key, val, found := strings.Cut(s, "=")
	if !found {
		return nil, ErrNotAVAlidTarget
	}

	sel := &metav1.LabelSelector{
		MatchLabels: map[string]string{
			key: val,
		},
	}

	return metav1.LabelSelectorAsSelector(sel)
}

func (t Target) String() string {
	return string(t)
}

func (t *Target) Set(s string) error {
	val := (*string)(t)
	*val = s
	return nil
}
