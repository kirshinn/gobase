package main

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

var tesOk = `1
2
2
3
3
3
4
4
5`

var tesOkResult = `1
2
3
4
5
`

func TestOk(t *testing.T) {
	in := bufio.NewReader(strings.NewReader(tesOk))
	out := new(bytes.Buffer)

	err := unique(in, out)
	if err != nil {
		t.Errorf("test for Ok Failed - error")
	}

	result := out.String()

	if result != tesOkResult {
		t.Errorf("test for Ok Failed - result not match\n %v %v", result, tesOkResult)
	}
}

var testFail = `1
2
1
`

func TestForError(t *testing.T) {
	in := bufio.NewReader(strings.NewReader(testFail))
	out := new(bytes.Buffer)

	err := unique(in, out)
	if err == nil {
		t.Errorf("test for Ok Failed - error %v", err)
	}
}
