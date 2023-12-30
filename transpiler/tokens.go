package transpiler

import (
	"fmt"
)

type tokenType int

type token struct {
	typ  tokenType
	lit  string
	line int
	col  int
}

const (
	// General tokens
	tEOF tokenType = iota
	tError
	tRoot
	tNewLine
	tPackage
	tImport
	tGoCode
	tHmltStart
	tHmltEnd

	// Hmlt/Haml tokens
	tDoctype
	tTag
	tId
	tClass
	tAttrName
	tAttrOperator
	tAttrStaticValue
	tAttrDynamicValue
	tIndent
	tStaticText
	tDynamicText
	tUnescaped
	tOutputCode
	tExecuteCode
	tRenderCommand
	tChildrenCommand
	tAttributesCommand
)

func (t tokenType) String() string {
	switch t {
	case tEOF:
		return "EOF"
	case tError:
		return "Error"
	case tRoot:
		return "Root"
	case tNewLine:
		return "NewLine"
	case tPackage:
		return "Package"
	case tImport:
		return "Import"
	case tGoCode:
		return "GoCode"
	case tHmltStart:
		return "HmltStart"
	case tHmltEnd:
		return "HmltEnd"
	case tDoctype:
		return "Doctype"
	case tTag:
		return "Tag"
	case tId:
		return "Id"
	case tClass:
		return "Class"
	case tAttrName:
		return "AttrName"
	case tAttrOperator:
		return "AttrOperator"
	case tAttrStaticValue:
		return "AttrStaticValue"
	case tAttrDynamicValue:
		return "AttrDynamicValue"
	case tIndent:
		return "Indent"
	case tStaticText:
		return "StaticText"
	case tDynamicText:
		return "DynamicText"
	case tUnescaped:
		return "Unescaped"
	case tOutputCode:
		return "OutputCode"
	case tExecuteCode:
		return "ExecuteCode"
	case tRenderCommand:
		return "RenderCommand"
	case tChildrenCommand:
		return "ChildrenCommand"
	case tAttributesCommand:
		return "AttributesCommand"
	default:
		return "!Unknown!"
	}
}

func (t token) Type() tokenType {
	return t.typ
}

func (t token) Lit() string {
	return t.lit
}

func (t token) Line() int {
	return t.line
}

func (t token) Col() int {
	return t.col
}

func (t token) String() string {
	if t.typ != tError && len(t.lit) > 30 {
		return fmt.Sprintf("%s[%d:%d]: %.30q...", t.typ, t.line, t.col, t.lit)
	}
	return fmt.Sprintf("%s[%d:%d]: %q", t.typ, t.line, t.col, t.lit)
}
