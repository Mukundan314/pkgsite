// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package godoc

import (
	"bytes"
	"go/format"
	"go/parser"
	"go/token"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/safehtml"
)

func TestParse(t *testing.T) {
	for _, test := range []struct {
		name, want string
		section    SectionType
	}{
		{
			name:    "sidenav",
			section: SidenavSection,
			want:    quoteSidenav,
		},
		{
			name:    "sidenav-mobile",
			section: SidenavMobileSection,
			want:    quoteSidenavMobile,
		},
		{
			name:    "body",
			section: BodySection,
			want:    quoteBody,
		},
	} {
		{
			t.Run(test.name, func(t *testing.T) {
				got, err := Parse(safehtml.HTMLEscaped(quoteDocHTML), test.section)
				if err != nil {
					t.Fatal(err)
				}
				if diff := cmp.Diff(test.want, got.String()); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}
			})
		}
	}
}

func TestRemoveUnusedASTNodes(t *testing.T) {
	const file = `
// Package-level comment.
package p

// const C
const C = 1

// leave unexported consts
const c = 1

// var V
var V int

// leave unexported vars
var v int

// type T
type T int

// leave unexported types
type t int

// Exp is exported.
func Exp() {}

// unexp is not exported.
func unexp() {}

// M is exported.
func (t T) M() int {}

// m isn't.
func (T) m() {}

// U is an exported method of an unexported type.
// Its doc is not shown, unless t is embedded
// in an exported type. We don't remove it to
// be safe.
func (t) U() {}
`
	////////////////
	const want = `// Package-level comment.
package p

// const C
const C = 1

// leave unexported consts
const c = 1

// var V
var V int

// leave unexported vars
var v int

// type T
type T int

// leave unexported types
type t int

// Exp is exported.
func Exp()

// M is exported.
func (t T) M() int

// U is an exported method of an unexported type.
// Its doc is not shown, unless t is embedded
// in an exported type. We don't remove it to
// be safe.
func (t) U()
`
	////////////////

	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, "tst.go", file, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	RemoveUnusedASTNodes(astFile)
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, astFile); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("mismatch (-want, +got):\n%s", diff)
	}
}
