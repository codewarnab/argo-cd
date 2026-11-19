package fixture

import (
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"golang.org/x/sync/errgroup"
)

var (
	matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")
)

// returns dns friends string which is no longer than 63 characters and has specified postfix at the end
func DnsFriendly(str string, postfix string) string { //nolint:revive //FIXME(var-naming)
	str = matchFirstCap.ReplaceAllString(str, "${1}-${2}")
	str = matchAllCap.ReplaceAllString(str, "${1}-${2}")
	str = strings.ToLower(str)
	str = strings.ReplaceAll(str, "/", "-")
	str = strings.ReplaceAll(str, "_", "-")

	if diff := len(str) + len(postfix) - 63; diff > 0 {
		str = str[:len(str)-diff]
	}
	return str + postfix
}

func RunFunctionsInParallelAndCheckErrors(t *testing.T, functions []func() error) {
	t.Helper()
	var eg errgroup.Group
	groupStart := time.Now()
	for _, function := range functions {
		f := runtime.FuncForPC(reflect.ValueOf(function).Pointer())
		file, line := f.FileLine(f.Entry())
		eg.Go(func() error {
			start := time.Now()
			err := function()
			fmt.Printf("E2E-TIMING step %s:%d %.3fs\n", file[strings.LastIndex(file, "/")+1:], line, time.Since(start).Seconds())
			return err
		})
	}
	require.NoError(t, eg.Wait())
	fmt.Printf("E2E-TIMING group %.3fs\n", time.Since(groupStart).Seconds())
}
