package internal

import (
	"go/ast"

	"github.com/vektra/mockery/v3/config"
	"golang.org/x/tools/go/packages"
)

type Interface struct {
	Name     string // Name of the type to be mocked.
	TypeSpec *ast.TypeSpec
	FileName string
	File     *ast.File
	Pkg      *packages.Package
	Config   *config.Config
}

func NewInterface(
	name string,
	typeSpec *ast.TypeSpec,
	filename string,
	file *ast.File,
	pkg *packages.Package,
	config *config.Config,
) *Interface {
	return &Interface{
		Name:     name,
		TypeSpec: typeSpec,
		FileName: filename,
		File:     file,
		Pkg:      pkg,
		Config:   config,
	}
}
