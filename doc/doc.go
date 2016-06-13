package doc

import (
	"strconv"
	"strings"

	"github.com/hashicorp/hcl/hcl/ast"
)

// Input represents a terraform input variable.
type Input struct {
	Name        string
	Description string
	Default     string
}

// Output represents a terraform output.
type Output struct {
	Name        string
	Description string
}

// Doc represents a terraform module doc.
type Doc struct {
	Comment string
	Inputs  []Input
	Outputs []Output
}

// Create creates a new *Doc from the supplied map
// of filenames and *ast.File.
func Create(files map[string]*ast.File) *Doc {
	doc := new(Doc)

	for _, f := range files {
		list := f.Node.(*ast.ObjectList)
		doc.Inputs = append(doc.Inputs, inputs(list)...)
		doc.Outputs = append(doc.Outputs, outputs(list)...)
	}

	return doc
}

// Inputs returns all variables from `list`.
func inputs(list *ast.ObjectList) []Input {
	var ret []Input

	for _, item := range list.Items {
		if is(item, "variable") {
			name, _ := strconv.Unquote(item.Keys[1].Token.Text)
			items := item.Val.(*ast.ObjectType).List.Items
			desc, _ := strconv.Unquote(get(items, "description"))
			def := get(items, "default")
			ret = append(ret, Input{
				Name:        name,
				Description: desc,
				Default:     def,
			})
		}
	}

	return ret
}

// Outputs returns all outputs from `list`.
func outputs(list *ast.ObjectList) []Output {
	var ret []Output

	for _, item := range list.Items {
		if is(item, "output") {
			name, _ := strconv.Unquote(item.Keys[1].Token.Text)

			var desc string
			if c := item.LeadComment; c != nil {
				desc = comment(c.List)
			}

			ret = append(ret, Output{
				Name:        name,
				Description: desc,
			})
		}
	}

	return ret
}

// Get `key` from the list of object `items`.
func get(items []*ast.ObjectItem, key string) string {
	for _, item := range items {
		if is(item, key) {
			if lit, ok := item.Val.(*ast.LiteralType); ok {
				return lit.Token.Text
			}

			return ""
		}
	}

	return ""
}

// Is returns true if `item` is of `kind`.
func is(item *ast.ObjectItem, kind string) bool {
	if len(item.Keys) > 0 {
		return item.Keys[0].Token.Text == kind
	}

	return false
}

// Unquote the given string.
func unquote(s string) string {
	s, _ = strconv.Unquote(s)
	return s
}

// Comment cleans and returns a comment.
func comment(l []*ast.Comment) string {
	var line string
	var ret string

	for _, t := range l {
		line = strings.TrimSpace(t.Text)
		line = strings.TrimPrefix(line, "//")
		ret += strings.TrimSpace(line) + "\n"
	}

	return ret
}
