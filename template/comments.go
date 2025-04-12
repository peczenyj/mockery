package template

import "go/ast"

type Comments struct {
	GenDeclDoc      CommentGroup
	TypeSpecComment CommentGroup
	TypeSpecDoc     CommentGroup
}

func NewComments(typeSpec *ast.TypeSpec, genDecl *ast.GenDecl) Comments {
	return Comments{
		GenDeclDoc:      NewCommentGroupFromAST(genDecl.Doc),
		TypeSpecComment: NewCommentGroupFromAST(typeSpec.Comment),
		TypeSpecDoc:     NewCommentGroupFromAST(typeSpec.Doc),
	}
}
