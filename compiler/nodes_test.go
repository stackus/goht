package compiler

import (
	"bytes"
	"testing"
)

func Test_Parsing(t *testing.T) {
	tests := map[string]struct {
		input   string
		want    string
		wantErr bool
	}{
		"full haml document": {
			input: `package testing

var foo = "bar"

@haml test(title string, err error) {
	!!! 5
	%html
		%head
			%title= title
		%body
			%p some text #{foo}
			#main-content
				%p= "Hello World"
			- if err != nil
				.error
					%p= "Something went wrong"
}
`,
			want: `Root
	GoCode
	Template
		Doctype
		NewLine
		Element html()
			NewLine
			Element head()
				NewLine
				Element title()
					Script
			Element body()
				NewLine
				Element p()
					Text(S)
					Text(D)
				Element div()
					NewLine
					Element p()
						Script
				SilentScript
					Element div()
						NewLine
						Element p()
							Script
	GoCode
`,
		},
		"full slim document": {
			input: `package testing

var foo = "bar"

@slim test(title string, err error) {
	doctype
	html
		head
			title= title
		body
			p some text #{foo}
			#main-content
				p= "Hello World"
			- if err != nil
				.error
					p= "Something went wrong"
}
`,
			want: `Root
	GoCode
	Template
		Doctype
		NewLine
		Element html()
			NewLine
			Element head()
				NewLine
				Element title()
					Script
			Element body()
				NewLine
				Element p()
					Text(S)
					Text(D)
				Element div()
					NewLine
					Element p()
						Script
				SilentScript
					Element div()
						NewLine
						Element p()
							Script
	GoCode
`,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			p := newParser([]byte(test.input))
			err := p.parse()
			if (err != nil) != test.wantErr {
				t.Fatalf("parse() error = %v, wantErr %v", err, test.wantErr)
			}
			buf := new(bytes.Buffer)
			_ = p.template.Root.Tree(buf, 0)
			got := buf.String()
			if got != test.want {
				t.Errorf("got \n%s----\n want \n%s----", got, test.want)
			}
		})
	}
}

func Test_RootNode(t *testing.T) {
	tests := map[string]struct {
		input   string
		want    string
		wantErr bool
	}{
		"empty": {
			input: "",
			want:  "Root\n",
		},
		"simple go": {
			input: "package simple\n\nvar simple = \"simple\"\n",
			want:  "Root\n\tGoCode\n",
		},
		"simple goht": {
			input: "package main\n@goht test() {\n}",
			want:  "Root\n\tGoCode\n\tTemplate\n",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			p := newParser([]byte(test.input))
			err := p.parse()
			if (err != nil) != test.wantErr {
				t.Fatalf("parse() error = %v, wantErr %v", err, test.wantErr)
			}
			buf := new(bytes.Buffer)
			_ = p.template.Root.Tree(buf, 0)
			got := buf.String()
			if got != test.want {
				t.Errorf("got \n%s----\n want \n%s----", got, test.want)
			}
		})
	}
}

func Test_TemplateNode(t *testing.T) {
	tests := map[string]struct {
		input   string
		want    string
		wantErr bool
	}{
		"empty": {
			input:   "@goht empty() {\n",
			want:    "Root\n\tTemplate\n",
			wantErr: true,
		},
		"simple": {
			input: "@goht test() {\n\tFoo\n}",
			want:  "Root\n\tTemplate\n\t\tText(S)\n\t\tNewLine\n",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			p := newParser([]byte(test.input))
			err := p.parse()
			if (err != nil) != test.wantErr {
				t.Fatalf("parse() error = %v, wantErr %v", err, test.wantErr)
			}
			buf := new(bytes.Buffer)
			_ = p.template.Root.Tree(buf, 0)
			got := buf.String()
			if got != test.want {
				t.Errorf("got \n%s----\nwant \n%s----", got, test.want)
			}
		})
	}
}

func Test_TextNode(t *testing.T) {
	tests := map[string]struct {
		input   string
		want    string
		wantErr bool
	}{
		"simple": {
			input: "@goht test() {\n\tFoo\n}",
			want:  "Root\n\tTemplate\n\t\tText(S)\n\t\tNewLine\n",
		},
		"with dynamic text": {
			input: "@goht test() {\n\tFoo #{foo}\n}",
			want:  "Root\n\tTemplate\n\t\tText(S)\n\t\tText(D)\n\t\tNewLine\n",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			p := newParser([]byte(test.input))
			err := p.parse()
			if (err != nil) != test.wantErr {
				t.Fatalf("parse() error = %v, wantErr %v", err, test.wantErr)
			}
			buf := new(bytes.Buffer)
			_ = p.template.Root.Tree(buf, 0)
			got := buf.String()
			if got != test.want {
				t.Errorf("got \n%s----\nwant \n%s----", got, test.want)
			}
		})
	}
}

func Test_UnescapeNode(t *testing.T) {
	tests := map[string]struct {
		input   string
		want    string
		wantErr bool
	}{
		"empty": {
			input: "@goht test() {\n\t!\n}",
			want: `Root
	Template
		Unescape
		NewLine
`,
		},
		"simple": {
			input: `@goht test() {
	! Foo
}`,
			want: `Root
	Template
		Unescape
			Text(S)
		NewLine
`,
		},
		"dynamic text": {
			input: "@goht test() {\n\t! #{foo}\n}",
			want:  "Root\n\tTemplate\n\t\tUnescape\n\t\t\tText(D)\n\t\tNewLine\n",
		},
		"static and dynamic text": {
			input: "@goht test() {\n\t! Foo #{foo}\n}",
			want:  "Root\n\tTemplate\n\t\tUnescape\n\t\t\tText(S)\n\t\t\tText(D)\n\t\tNewLine\n",
		},
		"illegal nesting": {
			input: `@goht test() {
	%p! foo
		bar
}`,
			want: `Root
	Template
		Element p()
			Unescape
				Text(S)
`,
			wantErr: true,
		},
		"mixed with tags": {
			input: `@goht test() {
	- var html = "<em>is</em>"
	%p This #{html} HTML.
	%p! This #{html} HTML.
}`,
			want: `Root
	Template
		SilentScript
		Element p()
			Text(S)
			Text(D)
			Text(S)
		Element p()
			Unescape
				Text(S)
				Text(D)
				Text(S)
`,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			p := newParser([]byte(test.input))
			err := p.parse()
			if (err != nil) != test.wantErr {
				t.Fatalf("parse() error = %v, wantErr %v", err, test.wantErr)
			}
			buf := new(bytes.Buffer)
			_ = p.template.Root.Tree(buf, 0)
			got := buf.String()
			if got != test.want {
				t.Errorf("got \n%s----\nwant \n%s----", got, test.want)
			}
		})
	}
}

func Test_ElementNode(t *testing.T) {
	tests := map[string]struct {
		input   string
		want    string
		wantErr bool
	}{
		"simple": {
			input: "@goht test() {\n\t%p\n}",
			want:  "Root\n\tTemplate\n\t\tElement p()\n\t\t\tNewLine\n",
		},
		"illegal nesting": {
			input: "@goht test() {\n\t%p foo\n\t\t%p bar\n}",
			want: `Root
	Template
		Element p()
			Text(S)
`,
			wantErr: true,
		},
		"unescaped text": {
			input: "@goht test() {\n\t%p! foo\n}",
			want: `Root
	Template
		Element p()
			Unescape
				Text(S)
`,
		},
		"unescaped text before new tag": {
			input: `@goht test() {
	%p! foo
	%p bar
}`,
			want: `Root
	Template
		Element p()
			Unescape
				Text(S)
		Element p()
			Text(S)
`,
		},
		"tags before and after void tag": {
			input: `@goht test() {
	%p#fizz.foo text
	%img{src: "foo.png"}
	%p#fizz.foo text
	%img{src: "foo.png"}
}`,
			want: `Root
	Template
		Element p()
			Text(S)
		Element img(src="foo.png")
			NewLine
		Element p()
			Text(S)
		Element img(src="foo.png")
			NewLine
`,
		},
		"tags before and after void tag character": {
			input: `@goht test() {
	%p#fizz.foo text
	%closed/
	%p#fizz.foo text
	%closed/
}`,
			want: `Root
	Template
		Element p()
			Text(S)
		Element closed()
			NewLine
		Element p()
			Text(S)
		Element closed()
			NewLine
`,
		},
		"with object ref": {
			input: "@goht test() {\n\t%p[foo]\n}",
			want:  "Root\n\tTemplate\n\t\tElement p()\n\t\t\tNewLine\n",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			p := newParser([]byte(test.input))
			err := p.parse()
			if (err != nil) != test.wantErr {
				t.Fatalf("parse() error = %v, wantErr %v", err, test.wantErr)
			}
			buf := new(bytes.Buffer)
			_ = p.template.Root.Tree(buf, 0)
			got := buf.String()
			if got != test.want {
				t.Errorf("got \n%s----\nwant \n%s----", got, test.want)
			}
		})
	}
}

func Test_ElementAttributes(t *testing.T) {
	tests := map[string]struct {
		input   string
		want    string
		wantErr bool
	}{
		"simple": {
			input: "@goht test() {\n\t%p{foo:\"bar\"}\n}",
			want: `Root
	Template
		Element p(foo="bar")
			NewLine
`,
		},
		"no tag": {
			input: "@goht test() {\n\t{foo:\"bar\"}\n}",
			want: `Root
	Template
		Element div(foo="bar")
			NewLine
`,
		},
		"dynamic attribute": {
			input: "@goht test() {\n\t%p{foo:#{bar}}\n}",
			want: `Root
	Template
		Element p(foo={bar})
			NewLine
`,
		},
		"quoted attribute names": {
			input: "@goht test() {\n\t%p{\"x:foo\":#{bar}, `@fizz`:`b\"uzz`}\n}",
			want: `Root
	Template
		Element p(x:foo={bar},@fizz="b\"uzz")
			NewLine
`,
		},
		"attributes command": {
			input: "@goht test() {\n\t%p{foo:#{bar}, @attributes:#{list}}\n}",
			want: `Root
	Template
		Element p(foo={bar},@attrs={list...})
			NewLine
`,
		},
		"multiline attributes": {
			input: `@goht test() {
	%p{
		foo:#{bar},
		@attributes:#{list}
	}
}`,
			want: `Root
	Template
		Element p(foo={bar},@attrs={list...})
			NewLine
`,
		},
		"indented elements": {
			input: `@goht test() {
	%p foo
	%p bar
}`,
			want: `Root
	Template
		Element p()
			Text(S)
		Element p()
			Text(S)
`,
		},
		"boolean attribute on tag with content": {
			input: `@goht test() {
	.foo{bar} fizz
}`,
			want: `Root
	Template
		Element div(bar)
			Text(S)
`,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			p := newParser([]byte(test.input))
			err := p.parse()
			if (err != nil) != test.wantErr {
				t.Fatalf("parse() error = %v, wantErr %v", err, test.wantErr)
			}
			buf := new(bytes.Buffer)
			_ = p.template.Root.Tree(buf, 0)
			got := buf.String()
			if got != test.want {
				t.Errorf("got \n%s----\nwant \n%s----", got, test.want)
			}
		})
	}
}

func Test_SilentScriptNode(t *testing.T) {
	tests := map[string]struct {
		input   string
		want    string
		wantErr bool
	}{
		"simple": {
			input: "@goht test() {\n\t- var foo = \"bar\"\n}",
			want: `Root
	Template
		SilentScript
`,
		},
		"nested content": {
			input: "@goht test() {\n\t- var foo = \"bar\"\n\t\t%p= foo\n}",
			want: `Root
	Template
		SilentScript
			Element p()
				Script
`,
		},
		"mixed indents": {
			input: "@goht test() {\n\t%p1 one\n\t- if foo {\n\t\t%p bar\n\t- }\n\t%p2 two\n}",
			want: `Root
	Template
		Element p1()
			Text(S)
		SilentScript
			Element p()
				Text(S)
		SilentScript
		Element p2()
			Text(S)
`,
		},
		"shorter indent": {
			input: "@goht test() {\n\t%p1\n\t\t- if foo\n\t%p2 two\n}",
			want: `Root
	Template
		Element p1()
			NewLine
			SilentScript
		Element p2()
			Text(S)
`,
		},
		"ruby style comment": {
			input: "@goht test() {\n\t-# foo\n\t%p bar\n}",
			want: `Root
	Template
		Element p()
			Text(S)
`,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			p := newParser([]byte(test.input))
			err := p.parse()
			if (err != nil) != test.wantErr {
				t.Fatalf("parse() error = %v, wantErr %v", err, test.wantErr)
			}
			buf := new(bytes.Buffer)
			_ = p.template.Root.Tree(buf, 0)
			got := buf.String()
			if got != test.want {
				t.Errorf("got \n%s----\nwant \n%s----", got, test.want)
			}
		})
	}
}

func Test_CommentNode(t *testing.T) {
	tests := map[string]struct {
		input   string
		want    string
		wantErr bool
	}{
		"same line": {
			input: "@goht test() {\n\t/ foo\n}",
			want: `Root
	Template
		Comment
			NewLine
`,
		},
		"nested content": {
			input: "@goht test() {\n\t/\n\t\t%p bar\n}",
			want: `Root
	Template
		Comment
			NewLine
			Element p()
				Text(S)
`,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			p := newParser([]byte(test.input))
			err := p.parse()
			if (err != nil) != test.wantErr {
				t.Fatalf("parse() error = %v, wantErr %v", err, test.wantErr)
			}
			buf := new(bytes.Buffer)
			_ = p.template.Root.Tree(buf, 0)
			got := buf.String()
			if got != test.want {
				t.Errorf("got \n%s----\nwant \n%s----", got, test.want)
			}
		})
	}
}

func Test_ScriptNode(t *testing.T) {
	tests := map[string]struct {
		input   string
		want    string
		wantErr bool
	}{
		"simple": {
			input: "@goht test() {\n\t= foo\n}",
			want: `Root
	Template
		Script
		NewLine
`,
		},
		"after element": {
			input: "@goht test() {\n\t%p= foo\n}",
			want: `Root
	Template
		Element p()
			Script
`,
		},
		"before content": {
			input: "@goht test() {\n\t= foo\n\t%p bar\n}",
			want: `Root
	Template
		Script
		NewLine
		Element p()
			Text(S)
`,
		},
		"mixed indents": {
			input: "@goht test() {\n\t%p1\n\t\t%p2= foo\n\t%p3 bar\n}",
			want: `Root
	Template
		Element p1()
			NewLine
			Element p2()
				Script
		Element p3()
			Text(S)
`,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			p := newParser([]byte(test.input))
			err := p.parse()
			if (err != nil) != test.wantErr {
				t.Fatalf("parse() error = %v, wantErr %v", err, test.wantErr)
			}
			buf := new(bytes.Buffer)
			_ = p.template.Root.Tree(buf, 0)
			got := buf.String()
			if got != test.want {
				t.Errorf("got \n%s----\nwant \n%s----", got, test.want)
			}
		})
	}
}

func Test_RenderNode(t *testing.T) {
	tests := map[string]struct {
		input   string
		want    string
		wantErr bool
	}{
		"simple": {
			input: "@goht test() {\n\t= @render foo()\n}",
			want: `Root
	Template
		RenderCommand
`,
		},
		"with children": {
			input: "@goht test() {\n\t= @render foo()\n\t\t%p bar\n}",
			want: `Root
	Template
		RenderCommand
			Element p()
				Text(S)
`,
		},
		"mixed indents": {
			input: "@goht test() {\n\t%p1 one\n\t= @render foo()\n\t\t%p bar\n\t%p2 two\n}",
			want: `Root
	Template
		Element p1()
			Text(S)
		RenderCommand
			Element p()
				Text(S)
		Element p2()
			Text(S)
`,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			p := newParser([]byte(test.input))
			err := p.parse()
			if (err != nil) != test.wantErr {
				t.Fatalf("parse() error = %v, wantErr %v", err, test.wantErr)
			}
			buf := new(bytes.Buffer)
			_ = p.template.Root.Tree(buf, 0)
			got := buf.String()
			if got != test.want {
				t.Errorf("got \n%s----\nwant \n%s----", got, test.want)
			}
		})
	}
}

func Test_ChildrenCommand(t *testing.T) {
	tests := map[string]struct {
		input   string
		want    string
		wantErr bool
	}{
		"simple": {
			input: "@goht test() {\n\t= @children\n}",
			want: `Root
	Template
		ChildrenCommand
`,
		},
		"mixed indents": {
			input: "@goht test() {\n\t%p1 one\n\t%parent\n\t\t= @children\n\t%p2 two\n}",
			want: `Root
	Template
		Element p1()
			Text(S)
		Element parent()
			NewLine
			ChildrenCommand
		Element p2()
			Text(S)
`,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			p := newParser([]byte(test.input))
			err := p.parse()
			if (err != nil) != test.wantErr {
				t.Fatalf("parse() error = %v, wantErr %v", err, test.wantErr)
			}
			buf := new(bytes.Buffer)
			_ = p.template.Root.Tree(buf, 0)
			got := buf.String()
			if got != test.want {
				t.Errorf("got \n%s----\nwant \n%s----", got, test.want)
			}
		})
	}
}
