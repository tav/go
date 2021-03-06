package p1

import (
	ptwo "p2"
)

const (
	ConstChase2 = constChase // forward declaration to unexported ident
	constChase  = AIsLowerA  // forward declaration to exported ident

	A         = 1
	a         = 11
	A64 int64 = 1

	AIsLowerA = a // previously declared
)

const (
	ConversionConst = MyInt(5)
)

// Variables from function calls.
var (
	V      = ptwo.F()
	VError = BarE()
	V1     = Bar1(1, 2, 3)
	V2     = ptwo.G()
)

var ChecksumError = ptwo.NewError("gzip checksum error")

const B = 2
const StrConst = "foo"
const FloatConst = 1.5

type myInt int

type MyInt int

type S struct {
	Public     *int
	private    *int
	PublicTime time.Time
}

type EmbedURLPtr struct {
	*url.URL
}

type S2 struct {
	S
	Extra bool
}

var X int64

var (
	Y int
	X I
)

type Namer interface {
	Name() string
}

type I interface {
	Namer
	ptwo.Twoer
	Set(name string, balance int64)
	Get(string) int64
	GetNamed(string) (balance int64)
	private()
}

type Error interface {
	error
	Temporary() bool
}

func (myInt) privateTypeMethod()           {}
func (myInt) CapitalMethodUnexportedType() {}

func (s *S2) SMethod(x int8, y int16, z int64) {}

type s struct{}

func (s) method()
func (s) Method()

func (S) StructValueMethod()
func (ignored S) StructValueMethodNamedRecv()

func (s *S2) unexported(x int8, y int16, z int64) {}

func Bar(x int8, y int16, z int64)                  {}
func Bar1(x int8, y int16, z int64) uint64          {}
func Bar2(x int8, y int16, z int64) (uint8, uint64) {}
func BarE() Error                                   {}

func unexported(x int8, y int16, z int64) {}

func TakesFunc(f func(dontWantName int) int)

type Codec struct {
	Func func(x int, y int) (z int)
}

type SI struct {
	I int
}

var SIVal = SI{}
var SIPtr = &SI{}
var SIPtr2 *SI

type T struct {
	common
}

type B struct {
	common
}

type common struct {
	i int
}

type TPtrUnexported struct {
	*common
}

type TPtrExported struct {
	*Embedded
}

type Embedded struct{}

func (*Embedded) OnEmbedded() {}

func (*T) JustOnT()             {}
func (*B) JustOnB()             {}
func (*common) OnBothTandBPtr() {}
func (common) OnBothTandBVal()  {}

type EmbedSelector struct {
	time.Time
}
