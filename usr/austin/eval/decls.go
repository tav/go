// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package eval

import (
	"bignum";
)

/*
 * Types
 */

type Value interface

type Type interface {
	// literal returns this type with all names recursively
	// stripped.  This should only be used when determining
	// assignment compatibility.  To strip a named type for use in
	// a type switch, use .rep().
	literal() Type;
	// rep returns the representative type.  If this is a named
	// type, this is the unnamed underlying type.  Otherwise, this
	// is an identity operation.
	rep() Type;
	// isBoolean returns true if this is a boolean type.
	isBoolean() bool;
	// isInteger returns true if this is an integer type.
	isInteger() bool;
	// isFloat returns true if this is a floating type.
	isFloat() bool;
	// isIdeal returns true if this is an ideal int or float.
	isIdeal() bool;
	// ZeroVal returns a new zero value of this type.
	Zero() Value;
	// String returns the string representation of this type.
	String() string;
}

type BoundedType interface {
	Type;
	// minVal returns the smallest value of this type.
	minVal() *bignum.Rational;
	// maxVal returns the largest value of this type.
	maxVal() *bignum.Rational;
}

/*
 * Values
 */

type Value interface {
	String() string;
	// Assign copies another value into this one.  It should
	// assume that the other value satisfies the same specific
	// value interface (BoolValue, etc.), but must not assume
	// anything about its specific type.
	Assign(o Value);
}

type BoolValue interface {
	Value;
	Get() bool;
	Set(bool);
}

type UintValue interface {
	Value;
	Get() uint64;
	Set(uint64);
}

type IntValue interface {
	Value;
	Get() int64;
	Set(int64);
}

type IdealIntValue interface {
	Value;
	Get() *bignum.Integer;
}

type FloatValue interface {
	Value;
	Get() float64;
	Set(float64);
}

type IdealFloatValue interface {
	Value;
	Get() *bignum.Rational;
}

type StringValue interface {
	Value;
	Get() string;
	Set(string);
}

type ArrayValue interface {
	Value;
	// TODO(austin) Get() is here for uniformity, but is
	// completely useless.  If a lot of other types have similarly
	// useless Get methods, just special-case these uses.
	Get() ArrayValue;
	Elem(i int64) Value;
}

type PtrValue interface {
	Value;
	Get() Value;
	Set(Value);
}

type Func interface
type FuncValue interface {
	Value;
	Get() Func;
	Set(Func);
}

/*
 * Scopes
 */

type Variable struct {
	// Index of this variable in the Frame structure
	Index int;
	// Static type of this variable
	Type Type;
}

type Constant struct {
	Type Type;
	Value Value;
}

// A definition can be a *Variable, *Constant, or Type.
type Def interface {}

type Scope struct {
	outer *Scope;
	defs map[string] Def;
	temps map[int] *Variable;
	numVars int;
	varTypes []Type;
}

func (s *Scope) Fork() *Scope
func (s *Scope) DefineVar(name string, t Type) *Variable
func (s *Scope) DefineTemp(t Type) *Variable
func (s *Scope) DefineConst(name string, t Type, v Value) *Constant
func (s *Scope) DefineType(name string, t Type) Type
func (s *Scope) Lookup(name string) (Def, *Scope)

// The universal scope
var universe = &Scope{defs: make(map[string] Def), temps: make(map[int] *Variable)};

/*
 * Frames
 */

type Frame struct {
	Outer *Frame;
	Scope *Scope;
	Vars []Value;
}

func (f *Frame) Get(s *Scope, index int) Value
func (f *Frame) String() string

func (s *Scope) NewFrame(outer *Frame) *Frame

/*
 * Functions
 */

type Func interface {
	NewFrame() *Frame;
	Call(*Frame);
}
