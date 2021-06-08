package caller_test

import (
	"runtime"
	"strings"
	"testing"

	"github.com/gdey/caller/simple/log"

	"github.com/gdey/caller"
)

func callCaller() runtime.Frame {
	var c caller.ACaller
	return c.Caller()
}

func TestCaller_Caller(t *testing.T) {
	type tcase struct {
		fn          func() runtime.Frame
		PrefixMatch bool
		Prefix      string
	}
	const expectedName = "github.com/gdey/caller_test.TestCaller_Caller"
	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			callerFrame := tc.fn()
			if tc.Prefix == "" {
				tc.Prefix = expectedName
			}
			if strings.HasPrefix(callerFrame.Function, tc.Prefix) != tc.PrefixMatch {
				t.Errorf("frame expected '%v' got '%v'", tc.Prefix, callerFrame.Function)
				return
			}
		}
	}

	tests := map[string]tcase{
		"from_this_file":            {fn: callCaller, PrefixMatch: true},
		"from_simple_log":           {fn: log.Caller, PrefixMatch: true},
		"not_in_ignore":             {fn: log.NotInIgnore},
		"ignore package":            {fn: log.Package, PrefixMatch: true},
		"ignore package and helper": {fn: log.PackageHelper, PrefixMatch: true},
		"ignore helper and package": {fn: log.HelperPackage, PrefixMatch: true},
		"ignore all": {
			fn: func() runtime.Frame {
				var c log.MyCaller
				c.Helper()
				return c.Caller()
			},
			PrefixMatch: true,
		},
		"ignore all no log": {
			fn: func() runtime.Frame {
				caller.SetNumberOfFramesToGet(20)
				_ = caller.NumberOfFramesToGet()
				caller.IgnorePackage()
				caller.Helper()
				return caller.Caller()
			},
			PrefixMatch: true,
			Prefix:      "testing.tRunner", // We are ignoring even this package, so it's going to be the end of the stack; which is the tRunner
		},
	}

	for name, tc := range tests {
		t.Run(name, fn(tc))
	}

}

func TestPackageName(t *testing.T) {
	fn := func(fullFunctionName, packageName string) (string, func(*testing.T)) {
		return fullFunctionName, func(t *testing.T) {
			pkg := caller.PackageName(fullFunctionName)
			if pkg != packageName {
				t.Errorf("package name, expected %v got %v", packageName, pkg)
			}
		}
	}
	tests := map[string]string{
		"github.com/gdey/caller_test.TestCaller_Caller.func1.10": "github.com/gdey/caller_test",
		"runtime.Caller":  "runtime",
		"Caller":          "",
		"gdey/caller.Foo": "gdey/caller",
	}
	for fnName, pkgName := range tests {
		t.Run(fn(fnName, pkgName))
	}
}
