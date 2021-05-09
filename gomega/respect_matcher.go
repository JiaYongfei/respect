package gomega

import (
	"fmt"
	"github.com/JiaYongfei/respect"
	"github.com/onsi/gomega/types"
)

func Respect(expected interface{}) types.GomegaMatcher {
	return &respectMatcher{
		expected: expected,
	}
}

type respectMatcher struct {
	expected interface{}
	diff     []string
}

func (matcher *respectMatcher) Match(actual interface{}) (success bool, err error) {
	matcher.diff = respect.Respect(actual, matcher.expected)
	return len(matcher.diff) == 0, nil
}

func (matcher *respectMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("%v", matcher.diff)
}

func (matcher *respectMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("%v", matcher.diff)
}
