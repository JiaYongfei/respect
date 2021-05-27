package gomega

import (
	"fmt"
	"github.com/JiaYongfei/respect"
	"github.com/onsi/gomega/types"
	"strings"
)

func Respect(expected interface{}, respectOptions ...respect.Options) types.GomegaMatcher {
	return &respectMatcher{
		expected: expected,
		options:  respectOptions,
	}
}

type respectMatcher struct {
	expected interface{}
	diff     []string
	options  []respect.Options
}

func (matcher *respectMatcher) Match(actual interface{}) (success bool, err error) {
	matcher.diff = respect.Respect(actual, matcher.expected, matcher.options...)
	return len(matcher.diff) == 0, nil
}

func (matcher *respectMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Diff:\n%v", strings.Join(matcher.diff, "\n"))
}

func (matcher *respectMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Diff:\n%v", strings.Join(matcher.diff, "\n"))
}
