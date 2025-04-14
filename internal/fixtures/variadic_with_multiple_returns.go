package test

type VariadicWithMultipleReturns interface {
	Foo(one string, two ...string) (result string, err error)
}
