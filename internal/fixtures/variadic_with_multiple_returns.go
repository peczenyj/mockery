package test

type VariadicWithMultipleReturns interface {
	Foo(one string, bar ...string) (result string, err error)
}
