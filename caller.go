// Package caller provides a helper to make getting caller information simpler, and adds the ability to ignore
// functions or packages, a la testing.Helper.
//
package caller

// This file contains the implementation of the caller helper functions and data structure.

import (
	"runtime"
	"strings"
)

const (
	// DefaultNumberOfFramesToGet is the number of frame we attempt to retrieve from the runtime to determine the
	// calling function. This value has been good enough for my tests, but may not be sufficient for very deep call
	// stacks. This value can be changed via the SetNumberOfFramesToGet function.
	DefaultNumberOfFramesToGet = 15
)

var defaultCaller ACaller

// PackageName will parse the full function name provided by a frame to find the package name
func PackageName(fullFuncName string) string {
	// we need to see if the name has a '/' in it; if so, we will need to
	// find the '.' after the last '/', if not then the first '.' is the
	// package separator
	slashIndex := strings.LastIndex(fullFuncName, "/")
	if slashIndex == -1 {
		slashIndex = 0 // use the whole string to find the '.'
	}
	// well we need to find the first '.' after the slashIndex
	dotIndex := strings.Index(fullFuncName[slashIndex:], ".")
	if dotIndex == -1 {
		// This is weird; we only have the function name, so we will return ""
		return ""
	}
	return fullFuncName[:slashIndex+dotIndex]
}

// ourPackage walks the frames to find our package name
func ourPackage() (packageName string) {
	var (
		frames = getFrames(4, 2)
		frame  runtime.Frame
		more   bool
	)

	// Loop to get frames.
	// A fixed number of pcs can expand to an indefinite number of Frames.
	for {
		frame, more = frames.Next()
		// this should be our package.
		// our package the first piece split on '.'
		packageName = PackageName(frame.Function)
		if packageName == "" && more {
			continue
		}
		break
	}
	return packageName
}

// getFrames will attempt retrieve the (num + skip) number of frames; then then skip passed the 'skip' number of frame.
func getFrames(num int, skip int) *runtime.Frames {
	// Ask runtime.Callers for up to 10 pcs, including runtime.Callers itself.
	pc := make([]uintptr, num+skip)
	n := runtime.Callers(0, pc)
	if n == 0 {
		// No pcs available. Stop now.
		// This can happen if the first argument to runtime.Callers is large.
		panic("no callers")
	}
	if skip >= (n - 1) {
		panic("not enough frames")
	}

	pc = pc[skip:n] // pass only valid pcs to runtime.CallersFrames
	return runtime.CallersFrames(pc)
}

//var ourPackageName = "github.com/gdey/caller"
var ourPackageName = ourPackage()

type ACaller struct {
	// numFramesToGet is the number of frame we should get; if this values is 0 or less it will default
	// to the default value
	numFramesToGet int
	// ignorePackages is the list of packages to ignore when walking the stack
	ignorePackages []string
	// ignoreFunctions is the list of functions to ignore when walking the stack
	ignoreFunctions []string
}

// IgnorePackage will mark the calling functions package as a package to ignore when
// the ACaller function is called in the search for the caller
//
// Note this should be called prior to any functions in the package calling the Helper() methods
// as that functions to it's ignore list who's package is not in the package list. This decreases the amount
// of memory used by the ACaller structure, as the Caller method always scans the the package ignore list first
// before scanning the function ignore list.
func (c *ACaller) IgnorePackage() {
	var (
		packageName string
		frames      = getFrames(5, 3)
		frame       runtime.Frame
		more        bool
	)
	for {
		frame, more = frames.Next()
		// this should be our package.
		// our package the first piece split on '.'
		packageName = PackageName(frame.Function)
		if packageName != "" || !more {
			if packageName == ourPackageName && more {
				// get the next frame only if there are more frames to get
				continue
			}
			break
		}
	}
	if packageName == "" {
		panic("Was not able to get the package name")
	}
	if packageName == ourPackageName || packageName == "runtime" {
		// Skip us or the runtime package
		return
	}
	c.ignorePackages = append(c.ignorePackages, packageName)
}

// Helper will mark the calling function as a function to ignore when
// the ACaller function is called in the search for the caller
//
// Use this only if there are only a few function that should be ignored
// If an entire package should be ignored call the IgnorePackage function
// instead. This function will not add the calling function if the calling
// functions package is already in the ignore list.
func (c *ACaller) Helper() {
	var (
		packageName string
		frames      = getFrames(5, 3)
		frame       runtime.Frame
		more        bool
	)
	for {
		frame, more = frames.Next()
		// this should be our package.
		// our package the first piece split on '.'
		if frame.Function == "" {
			if !more {
				panic("Was not able to get the function name ran out of frames")
			}
			continue
		}
		packageName = PackageName(frame.Function)

		if packageName != ourPackageName && packageName != "runtime" {
			// This check should not be necessary; as we skip the first three frames, which should be the
			// runtime.Caller
			// github.com/gdey/caller.getFrames
			// github.com/gdey/caller.Helper
			//
			// But we do this just in case the order changes
			break // we found the package
		}
		if !more {
			return // there is no frames, so return the package
		}
	}
	// Let's make sure we don't already have this in our ignore list
	for _, fnName := range c.ignoreFunctions {
		if frame.Function == fnName {
			return // already have it in out list
		}
	}
	// Let's make sure the package is not already ignored; if it is;
	// then we don't need to add this function
	if packageName == ourPackageName || packageName == "runtime" {
		// Skip us or the runtime package
		return
	}
	for _, pkgName := range c.ignorePackages {
		if packageName == pkgName {
			// skip adding it to our list as the package is already in our list
			return
		}
	}
	c.ignoreFunctions = append(c.ignoreFunctions, frame.Function)
}

// IgnoreFunction will mark the named function in the callers package as a function to ignore when
// the ACaller function is called in the search for the caller
//
// Use this only if there are only a few function that should be ignored
// If an entire package should be ignored call the IgnorePackage function
// instead. This function will not add the calling function if the calling
// functions package is already in the ignore list.
func (c *ACaller) IgnoreFunction(name string) {
	var (
		packageName string
		frames      = getFrames(5, 3)
		frame       runtime.Frame
		more        bool
	)
	for {
		frame, more = frames.Next()
		// this should be our package.
		// our package the first piece split on '.'
		if frame.Function == "" {
			if !more {
				panic("Was not able to get the function name ran out of frames")
			}
			continue
		}
		packageName = PackageName(frame.Function)

		if packageName != ourPackageName && packageName != "runtime" {
			// This check should not be necessary; as we skip the first three frames, which should be the
			// runtime.Caller
			// github.com/gdey/caller.getFrames
			// github.com/gdey/caller.Helper
			//
			// But we do this just in case the order changes
			break // we found the package
		}
		if !more {
			return // there is no frames, so return the package
		}
	}
	fullFunctionName := packageName + "." + name
	// Let's make sure we don't already have this in our ignore list
	for _, fnName := range c.ignoreFunctions {
		if fullFunctionName == fnName {
			return // already have it in out list
		}
	}
	// Let's make sure the package is not already ignored; if it is;
	// then we don't need to add this function
	if packageName == ourPackageName || packageName == "runtime" {
		// Skip us or the runtime package
		return
	}
	for _, pkgName := range c.ignorePackages {
		if packageName == pkgName {
			// skip adding it to our list as the package is already in our list
			return
		}
	}
	c.ignoreFunctions = append(c.ignoreFunctions, fullFunctionName)
}

// skipFrame will return weather the given frame is in one of the
// ignore lists
func (c *ACaller) skipFrame(frame runtime.Frame) bool {
	functionName := frame.Function
	packageName := PackageName(functionName)
	// We always skip runtime and this package
	if packageName == "runtime" || packageName == ourPackageName {
		return true
	}
	// go through the packages first
	for _, pkgName := range c.ignorePackages {
		if packageName == pkgName {
			// skip adding it to our list
			return true
		}
	}
	// go through the functions next.
	for _, fnName := range c.ignoreFunctions {
		if functionName == fnName {
			return true
		}
	}
	return false
}

// SetNumberOfFramesToGet will change the default number of frame to get.
func (c *ACaller) SetNumberOfFramesToGet(size uint) {
	if size > DefaultNumberOfFramesToGet {
		c.numFramesToGet = int(size)
	}
}

// NumberOfFramesToGet returns the number of frames we will we get from the runtime
func (c ACaller) NumberOfFramesToGet() int {
	if c.numFramesToGet != 0 {
		return c.numFramesToGet
	}
	return DefaultNumberOfFramesToGet
}

// Caller will walk up the call stack to find the caller that lead to the call of the function
// that called Caller. It will ignore any caller in the frame that is in it's ignore lists.
func (c ACaller) Caller() (frame runtime.Frame) {
	var more bool

	frames := getFrames(c.NumberOfFramesToGet(), 4)
	for {
		frame, more = frames.Next()
		if !c.skipFrame(frame) || !more {
			// we will return the last frame. (It is possible that out size is not big enough)
			break
		}
	}
	return frame
}

// Caller will walk up the call stack to find the caller that lead to the call of this function. It will ignore any callers
// in the frame that is in the ignore lists.
func Caller() (frame runtime.Frame) { return defaultCaller.Caller() }

// Helper will add the calling function to the function ignore list
func Helper() { defaultCaller.Helper() }

// IgnoreFunction is a more efficient was to add frequently called functions to the ignore list.
func IgnoreFunction(name string) { defaultCaller.IgnoreFunction(name) }

// IgnorePackage will add the package of the calling function to the packages ignore list
func IgnorePackage() { defaultCaller.IgnorePackage() }

// SetNumberOfFramesToGet will change the default number of frames to retrieve. This should not changed unless you know
// the it needs to be changed
func SetNumberOfFramesToGet(size uint) { defaultCaller.SetNumberOfFramesToGet(size) }

// NumberOfFramesToGet will return the current configured number of frames to get
func NumberOfFramesToGet() int { return defaultCaller.NumberOfFramesToGet() }

// Copyright 2021 Gautam Dey. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE FILE
