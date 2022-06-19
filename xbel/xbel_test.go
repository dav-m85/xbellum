package xbel

import (
	"testing"

	"github.com/matryer/is"
)

func TestDiff(t *testing.T) {
	is := is.New(t)

	a := []*Bookmark{
		{Href: "foo"}, {Href: "bar"}, {Href: "plop"},
	}
	b := []*Bookmark{
		{Href: "foo"}, {Href: "baz"}, {Href: "plop"},
	}
	onlyA, onlyB := Diff(a, b)
	is.True(len(onlyA) == 1)
	is.True(len(onlyB) == 1)
	is.True(onlyA[0].Href == "bar")
	is.True(onlyB[0].Href == "baz")
}
