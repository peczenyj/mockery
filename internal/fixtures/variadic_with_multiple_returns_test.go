package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNoUnrollVariadic(t *testing.T) {
	m := NewMockVariadicWithMultipleReturns(t)
	m.EXPECT().Foo(mock.Anything, mock.Anything).RunAndReturn(
		func(one string, two ...string) (string, error) {
			var s string = one
			for _, t := range two {
				s += t
			}
			return s, nil
		},
	)
	ret, err := m.Foo("one", "two", "three")
	assert.NoError(t, err)
	assert.Equal(t, "onetwothree", ret)
}

func TestUnrollVariadic(t *testing.T) {
	m := NewMockVariadicWithMultipleReturnsUnrollVariadic(t)
	m.EXPECT().Foo(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(
		func(one string, two ...string) (string, error) {
			var s string = one
			for _, t := range two {
				s += t
			}
			return s, nil
		},
	)
	ret, err := m.Foo("one", "two", "three")
	assert.NoError(t, err)
	assert.Equal(t, "onetwothree", ret)
}
