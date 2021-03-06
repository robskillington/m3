// Copyright (c) 2019 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/mauricelam/genny

package context

import (
	"github.com/m3db/m3/src/x/pool"
)

// Copyright (c) 2018 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// finalizeablesArrayPool provides a pool for finalizeable slices.
type finalizeablesArrayPool interface {
	// Init initializes the array pool, it needs to be called
	// before Get/Put use.
	Init()

	// Get returns the a slice from the pool.
	Get() []finalizeable

	// Put returns the provided slice to the pool.
	Put(elems []finalizeable)
}

type finalizeablesFinalizeFn func([]finalizeable) []finalizeable

type finalizeablesArrayPoolOpts struct {
	Options     pool.ObjectPoolOptions
	Capacity    int
	MaxCapacity int
	FinalizeFn  finalizeablesFinalizeFn
}

type finalizeablesArrPool struct {
	opts finalizeablesArrayPoolOpts
	pool pool.ObjectPool
}

func newFinalizeablesArrayPool(opts finalizeablesArrayPoolOpts) finalizeablesArrayPool {
	if opts.FinalizeFn == nil {
		opts.FinalizeFn = defaultFinalizeablesFinalizerFn
	}
	p := pool.NewObjectPool(opts.Options)
	return &finalizeablesArrPool{opts, p}
}

func (p *finalizeablesArrPool) Init() {
	p.pool.Init(func() interface{} {
		return make([]finalizeable, 0, p.opts.Capacity)
	})
}

func (p *finalizeablesArrPool) Get() []finalizeable {
	return p.pool.Get().([]finalizeable)
}

func (p *finalizeablesArrPool) Put(arr []finalizeable) {
	arr = p.opts.FinalizeFn(arr)
	if max := p.opts.MaxCapacity; max > 0 && cap(arr) > max {
		return
	}
	p.pool.Put(arr)
}

func defaultFinalizeablesFinalizerFn(elems []finalizeable) []finalizeable {
	var empty finalizeable
	for i := range elems {
		elems[i] = empty
	}
	elems = elems[:0]
	return elems
}

type finalizeablesArr []finalizeable

func (elems finalizeablesArr) grow(n int) []finalizeable {
	if cap(elems) < n {
		elems = make([]finalizeable, n)
	}
	elems = elems[:n]
	// following compiler optimized memcpy impl
	// https://github.com/golang/go/wiki/CompilerOptimizations#optimized-memclr
	var empty finalizeable
	for i := range elems {
		elems[i] = empty
	}
	return elems
}
