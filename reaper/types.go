package reaper

import (
	"errors"
	"flag"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ErrNotAVAlidTarget = errors.New("not a valid target")

var _ flag.Value = (*Target)(nil)

type Target string

func (t Target) Selector() (*metav1.LabelSelector, error) {
	s := string(t)
	if s == "" {
		return nil, ErrNotAVAlidTarget
	}

	key, val, found := strings.Cut(s, "=")
	if !found {
		return nil, ErrNotAVAlidTarget
	}

	return &metav1.LabelSelector{
		MatchLabels: map[string]string{
			key: val,
		},
	}, nil
}

func (t Target) String() string {
	return string(t)
}

func (t *Target) Set(s string) error {
	val := (*string)(t)
	*val = s
	return nil
}
