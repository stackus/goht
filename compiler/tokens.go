package compiler

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

	// Template tokens
	tTemplateStart
	tTemplateEnd

	// Goht/Haml tokens
	tDoctype
	tTag
	tId
	tClass
	tObjectRef
	tAttrName
	tAttrOperator
	tAttrEscapedValue
	tAttrDynamicValue
	tIndent
	tComment
	tRubyComment
	tVoidTag
	tNukeInnerWhitespace
	tNukeOuterWhitespace
	tEscapedText
	tDynamicText
	tPlainText
	tPreserveText
	tUnescaped
	tScript
	tSilentScript
	tRenderCommand
	tChildrenCommand
	tAttributesCommand
	tFilterStart
	tFilterEnd
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
	// case tGohtStart:
	// 	return "GohtStart"
	// case tGohtEnd:
	// 	return "GohtEnd"
	case tTemplateStart:
		return "TemplateStart"
	case tTemplateEnd:
		return "TemplateEnd"
	case tDoctype:
		return "Doctype"
	case tTag:
		return "Tag"
	case tId:
		return "Id"
	case tClass:
		return "Class"
	case tObjectRef:
		return "ObjectRef"
	case tAttrName:
		return "AttrName"
	case tAttrOperator:
		return "AttrOperator"
	case tAttrEscapedValue:
		return "AttrEscapedValue"
	case tAttrDynamicValue:
		return "AttrDynamicValue"
	case tIndent:
		return "Indent"
	case tComment:
		return "Comment"
	case tRubyComment:
		return "RubyComment"
	case tVoidTag:
		return "VoidTag"
	case tNukeInnerWhitespace:
		return "NukeInnerWhitespace"
	case tNukeOuterWhitespace:
		return "NukeOuterWhitespace"
	case tEscapedText:
		return "EscapedText"
	case tDynamicText:
		return "DynamicText"
	case tPlainText:
		return "PlainText"
	case tPreserveText:
		return "PreserveText"
	case tUnescaped:
		return "Unescaped"
	case tScript:
		return "Script"
	case tSilentScript:
		return "SilentScript"
	case tRenderCommand:
		return "RenderCommand"
	case tChildrenCommand:
		return "ChildrenCommand"
	case tAttributesCommand:
		return "AttributesCommand"
	case tFilterStart:
		return "FilterStart"
	case tFilterEnd:
		return "FilterEnd"
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
		return fmt.Sprintf("%s[%d:%d]: %q", t.typ, t.line, t.col, string([]rune(t.lit)[:30])+"...")
	}
	return fmt.Sprintf("%s[%d:%d]: %q", t.typ, t.line, t.col, t.lit)
}
