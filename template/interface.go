package template

import (
	"go/ast"

	"github.com/vektra/mockery/v3/template_funcs"
)

// Interface is the data used to generate a mock for some interface.
type Interface struct {
	// Comment contains line comments, if any.
	Comment CommentGroup
	// Doc contains the associated documentation, if any.
	Doc     CommentGroup
	Methods []Method
	// Name is the name of the original interface.
	Name string
	// StructName is the chosen name for the struct that will implement the interface.
	StructName   string
	TemplateData TemplateData
	TypeParams   []TypeParam
}

func NewInterface(
	name string,
	structName string,
	typeParams []TypeParam,
	methods []Method,
	templateData TemplateData,
	comment *ast.CommentGroup,
	doc *ast.CommentGroup,
) Interface {
	var commentGroup, docGroup CommentGroup
	if comment != nil {
		commentGroup.Text = comment.Text()
		commentGroup.List = []Comment{}
		for _, comment := range comment.List {
			commentGroup.List = append(commentGroup.List, Comment(comment.Text))
		}
	}
	if doc != nil {
		docGroup.Text = doc.Text()
		docGroup.List = []Comment{}
		for _, comment := range doc.List {
			docGroup.List = append(docGroup.List, Comment(comment.Text))
		}
	}
	return Interface{
		Name:         name,
		StructName:   structName,
		TypeParams:   typeParams,
		Methods:      methods,
		TemplateData: templateData,
		Comment:      commentGroup,
		Doc:          docGroup,
	}
}

func (m Interface) TypeConstraintTest() string {
	if len(m.TypeParams) == 0 {
		return ""
	}
	s := "["
	for idx, param := range m.TypeParams {
		if idx != 0 {
			s += ", "
		}
		s += template_funcs.Exported(param.Name())
		s += " "
		s += param.TypeString()
	}
	s += "]"
	return s
}

func (m Interface) TypeConstraint() string {
	if len(m.TypeParams) == 0 {
		return ""
	}
	s := "["
	for idx, param := range m.TypeParams {
		if idx != 0 {
			s += ", "
		}
		s += template_funcs.Exported(param.Name())
		s += " "
		s += param.TypeString()
	}
	s += "]"
	return s
}

func (m Interface) TypeInstantiation() string {
	if len(m.TypeParams) == 0 {
		return ""
	}
	s := "["
	for idx, param := range m.TypeParams {
		if idx != 0 {
			s += ", "
		}
		s += template_funcs.Exported(param.Name())
	}
	s += "]"
	return s
}
