package compiler

import (
	"testing"
)

func Test_TokenStringer(t *testing.T) {
	tests := map[string]struct {
		typ  tokenType
		want string
	}{
		"tEOF": {
			typ:  tEOF,
			want: "EOF",
		},
		"tError": {
			typ:  tError,
			want: "Error",
		},
		"tRoot": {
			typ:  tRoot,
			want: "Root",
		},
		"tNewLine": {
			typ:  tNewLine,
			want: "NewLine",
		},
		"tPackage": {
			typ:  tPackage,
			want: "Package",
		},
		"tImport": {
			typ:  tImport,
			want: "Import",
		},
		"tGoCode": {
			typ:  tGoCode,
			want: "GoCode",
		},
		"tGohtStart": {
			typ:  tGohtStart,
			want: "GohtStart",
		},
		"tGohtEnd": {
			typ:  tGohtEnd,
			want: "GohtEnd",
		},
		"tDoctype": {
			typ:  tDoctype,
			want: "Doctype",
		},
		"tTag": {
			typ:  tTag,
			want: "Tag",
		},
		"tId": {
			typ:  tId,
			want: "Id",
		},
		"tClass": {
			typ:  tClass,
			want: "Class",
		},
		"tObjectRef": {
			typ:  tObjectRef,
			want: "ObjectRef",
		},
		"tAttrName": {
			typ:  tAttrName,
			want: "AttrName",
		},
		"tAttrOperator": {
			typ:  tAttrOperator,
			want: "AttrOperator",
		},
		"tAttrEscapedValue": {
			typ:  tAttrEscapedValue,
			want: "AttrEscapedValue",
		},
		"tAttrDynamicValue": {
			typ:  tAttrDynamicValue,
			want: "AttrDynamicValue",
		},
		"tIndent": {
			typ:  tIndent,
			want: "Indent",
		},
		"tComment": {
			typ:  tComment,
			want: "Comment",
		},
		"tRubyComment": {
			typ:  tRubyComment,
			want: "RubyComment",
		},
		"tVoidTag": {
			typ:  tVoidTag,
			want: "VoidTag",
		},
		"tNukeInnerWhitespace": {
			typ:  tNukeInnerWhitespace,
			want: "NukeInnerWhitespace",
		},
		"tNukeOuterWhitespace": {
			typ:  tNukeOuterWhitespace,
			want: "NukeOuterWhitespace",
		},
		"tEscapedText": {
			typ:  tEscapedText,
			want: "EscapedText",
		},
		"tDynamicText": {
			typ:  tDynamicText,
			want: "DynamicText",
		},
		"tPlainText": {
			typ:  tPlainText,
			want: "PlainText",
		},
		"tPreserveText": {
			typ:  tPreserveText,
			want: "PreserveText",
		},
		"tUnescaped": {
			typ:  tUnescaped,
			want: "Unescaped",
		},
		"tScript": {
			typ:  tScript,
			want: "Script",
		},
		"tSilentScript": {
			typ:  tSilentScript,
			want: "SilentScript",
		},
		"tRenderCommand": {
			typ:  tRenderCommand,
			want: "RenderCommand",
		},
		"tChildrenCommand": {
			typ:  tChildrenCommand,
			want: "ChildrenCommand",
		},
		"tAttributesCommand": {
			typ:  tAttributesCommand,
			want: "AttributesCommand",
		},
		"tFilterStart": {
			typ:  tFilterStart,
			want: "FilterStart",
		},
		"tFilterEnd": {
			typ:  tFilterEnd,
			want: "FilterEnd",
		},
		"tUnknown": {
			typ:  tokenType(999),
			want: "!Unknown!",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.typ.String()
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func Test_TokenString(t *testing.T) {
	tests := map[string]struct {
		token token
		want  string
	}{
		"short": {
			token: token{typ: tGoCode, lit: "short", line: 0, col: 0},
			want:  "GoCode[0:0]: \"short\"",
		},
		"long": {
			token: token{typ: tGoCode, lit: "func HomePage(props HomePageProps) {", line: 10, col: 20},
			want:  "GoCode[10:20]: \"func HomePage(props HomePagePr...\"",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.token.String()
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
