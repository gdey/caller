package log

import (
	"runtime"

	"github.com/gdey/caller"
)

type MyCaller struct {
	caller.ACaller
}

func (c MyCaller) notInIgnore() runtime.Frame { return c.Caller() }
func (c MyCaller) Double() runtime.Frame      { return c.notInIgnore() }

func Caller() runtime.Frame {
	var c MyCaller
	c.Helper()
	return c.notInIgnore()
}

func NotInIgnore() runtime.Frame {
	var c MyCaller
	return c.notInIgnore()
}

func Package() runtime.Frame {
	var c MyCaller
	c.IgnorePackage()
	return c.Double()
}
func PackageHelper() runtime.Frame {
	var c MyCaller
	// tell the system to ignore the whole package
	c.IgnorePackage()
	// tell it to ignore this function, this should not get added to the list of ignored functions
	// because we are already ignoring the package
	c.Helper()
	return c.Double()
}

func HelperPackage() runtime.Frame {
	var c MyCaller
	// tell it to ignore this function, this will get added, because we are not ignoring the package at this time
	c.Helper()
	// tell the system to ignore the whole package; but will not remove the function from the list of functions to
	// ignore.
	// TODO(gdey) : should this be the case? Or should we remove functions that are no longer going to be looked for
	// and clean up list of functions. This makes IgnorePackage more complicated and time consuming; but as you
	// only call it once -- so might make sense to take the hit.
	c.IgnorePackage()
	return c.Double()
}
