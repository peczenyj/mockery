package test

import (
	"testing"

	"github.com/stretchr/testify/mock"
)

func TestDoer(t *testing.T) {
	d := NewMockNilRun(t)
	d.EXPECT().Foo(mock.Anything).Run(func(_ NilRun) {})
	d.Foo(nil)
}
