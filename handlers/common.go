package handlers

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	gomock "github.com/golang/mock/gomock"
)

// these form vars are not regular input fields, but transmit meta form info
var specialFormVars = map[string]bool{
	"save-and-return": true,
	":uri":            true,
	"q":               true,
}

// getOptionsAndRedirect iterates the provided form values and creates a list of options
// and updates a redirectURI if the form contains a redirect.
func getOptionsAndRedirect(form url.Values, redirectURI *string) (options []string) {
	options = []string{}
	for k := range form {
		if _, foundSpecial := specialFormVars[k]; foundSpecial {
			continue
		}

		if strings.Contains(k, "redirect:") {
			redirectReg := regexp.MustCompile(`^redirect:(.+)$`)
			redirectSubs := redirectReg.FindStringSubmatch(k)
			*redirectURI = redirectSubs[1]
			continue
		}

		options = append(options, k)
	}
	return options
}

// go-mock tailored matcher to compare lists of strings ignoring order
type itemsEq struct{ expected []string }

// ItemsEq checks if 2 slices contain the same items in any order
func ItemsEq(expected []string) gomock.Matcher {
	return &itemsEq{expected}
}

func (i *itemsEq) Matches(x interface{}) bool {
	if len(x.([]string)) != len(i.expected) {
		return false
	}
	mExpected := make(map[string]struct{})
	for _, e := range i.expected {
		mExpected[e] = struct{}{}
	}
	for _, val := range x.([]string) {
		if _, found := mExpected[val]; !found {
			return false
		}
	}
	return true
}

func (i *itemsEq) String() string {
	return fmt.Sprintf("%v (in any order)", i.expected)
}
