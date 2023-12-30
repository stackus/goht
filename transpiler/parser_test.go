package transpiler

import (
	"bytes"
	"testing"
)

func Test_HmltParsing(t *testing.T) {
	tests := map[string]struct {
		input   string
		want    string
		wantErr bool
	}{
		"full document": {
			input: `package testing

var foo = "bar"

@hmlt test(title string, err error) {
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
	Hmlt
		Doctype
		Element html()
			Element head()
				Element title()
					OutputCode
			Element body()
				Element p()
					Text(S)
					Text(D)
				Element div()
					Element p()
						OutputCode
				ExecuteCode
					Element div()
						Element p()
							OutputCode
	GoCode
`,
		},
		"with doctype": {
			input: `@hmlt basic3(fizz string) {
!!! 5
%html
	%head
		%title= {fizz}
	%body
		%p= {fizz}
}`,
			want: `Root
	Hmlt
		Doctype
		Element html()
			Element head()
				Element title()
					OutputCode
			Element body()
				Element p()
					OutputCode
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
		"simple hmlt": {
			input: "package main\n@hmlt test() {\n}",
			want:  "Root\n\tGoCode\n\tHmlt\n",
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

func Test_HmltNode(t *testing.T) {
	tests := map[string]struct {
		input   string
		want    string
		wantErr bool
	}{
		"empty": {
			input:   "@hmlt empty() {\n",
			want:    "Root\n\tHmlt\n",
			wantErr: true,
		},
		"simple": {
			input: "@hmlt test() {\nFoo\n}",
			want:  "Root\n\tHmlt\n\t\tText(S)\n",
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
			input: "@hmlt test() {\nFoo\n}",
			want:  "Root\n\tHmlt\n\t\tText(S)\n",
		},
		"with dynamic text": {
			input: "@hmlt test() {\nFoo #{foo}\n}",
			want:  "Root\n\tHmlt\n\t\tText(S)\n\t\tText(D)\n",
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
			input: "@hmlt test() {\n!\n}",
			want:  "Root\n\tHmlt\n\t\tUnescape\n",
		},
		"simple": {
			input: "@hmlt test() {\n! Foo\n}",
			want:  "Root\n\tHmlt\n\t\tUnescape\n\t\t\tText(S)\n",
		},
		"dynamic text": {
			input: "@hmlt test() {\n! #{foo}\n}",
			want:  "Root\n\tHmlt\n\t\tUnescape\n\t\t\tText(D)\n",
		},
		"static and dynamic text": {
			input: "@hmlt test() {\n! Foo #{foo}\n}",
			want:  "Root\n\tHmlt\n\t\tUnescape\n\t\t\tText(S)\n\t\t\tText(D)\n",
		},
		"illegal nesting": {
			input: "@hmlt test() {\n%p! foo\n\tbar\n}",
			want: `Root
	Hmlt
		Element p()
			Unescape
				Text(S)
`,
			wantErr: true,
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
			input: "@hmlt test() {\n%p\n}",
			want:  "Root\n\tHmlt\n\t\tElement p()\n",
		},
		"illegal nesting": {
			input: "@hmlt test() {\n%p foo\n\t%p bar\n}",
			want: `Root
	Hmlt
		Element p()
			Text(S)
`,
			wantErr: true,
		},
		"unescaped text": {
			input: "@hmlt test() {\n%p! foo\n}",
			want: `Root
	Hmlt
		Element p()
			Unescape
				Text(S)
`,
		},
		"unescaped text before new tag": {
			input: `@hmlt test() {
%p! foo
%p bar
}`,
			want: `Root
	Hmlt
		Element p()
			Unescape
				Text(S)
		Element p()
			Text(S)
`,
		},
		"illegal nesting with void tag": {
			input: `@hmlt test() {
	%p#fizz.foo text
	%img{src: "foo.png"}
	%p#fizz.foo text
	%img{src: "foo.png"}
}`,
			want: `Root
	Hmlt
		Element p()
			Text(S)
		Element img(src="foo.png")
		Element p()
			Text(S)
		Element img(src="foo.png")
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

func Test_ElementAttributes(t *testing.T) {
	tests := map[string]struct {
		input   string
		want    string
		wantErr bool
	}{
		"simple": {
			input: "@hmlt test() {\n%p{foo:\"bar\"}\n}",
			want: `Root
	Hmlt
		Element p(foo="bar")
`,
		},
		"dynamic attribute": {
			input: "@hmlt test() {\n%p{foo:{bar}}\n}",
			want: `Root
	Hmlt
		Element p(foo={bar})
`,
		},
		"quoted attribute names": {
			input: "@hmlt test() {\n%p{\"x:foo\":{bar}, `@fizz`:`b\"uzz`}\n}",
			want: `Root
	Hmlt
		Element p(x:foo={bar},@fizz="b\"uzz")
`,
		},
		"attributes command": {
			input: "@hmlt test() {\n%p{foo:{bar}, @attributes:{list}}\n}",
			want: `Root
	Hmlt
		Element p(foo={bar},@attrs={list...})
`,
		},
		"multiline attributes": {
			input: `@hmlt test() {
%p{
	foo:{bar},
	@attributes:{list}
}
}`,
			want: `Root
	Hmlt
		Element p(foo={bar},@attrs={list...})
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

func Test_ExecuteCodeNode(t *testing.T) {
	tests := map[string]struct {
		input   string
		want    string
		wantErr bool
	}{
		"simple": {
			input: "@hmlt test() {\n- var foo = \"bar\"\n}",
			want: `Root
	Hmlt
		ExecuteCode
`,
		},
		"nested content": {
			input: "@hmlt test() {\n- var foo = \"bar\"\n\t%p= foo\n}",
			want: `Root
	Hmlt
		ExecuteCode
			Element p()
				OutputCode
`,
		},
		"mixed indents": {
			input: "@hmlt test() {\n%p1 one\n- if foo\n\t%p bar\n%p2 two\n}",
			want: `Root
	Hmlt
		Element p1()
			Text(S)
		ExecuteCode
			Element p()
				Text(S)
		Element p2()
			Text(S)
`,
		},
		"shorter indent": {
			input: "@hmlt test() {\n%p1\n\t- if foo\n%p2 two\n}",
			want: `Root
	Hmlt
		Element p1()
			ExecuteCode
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

func Test_OutputCodeNode(t *testing.T) {
	tests := map[string]struct {
		input   string
		want    string
		wantErr bool
	}{
		"simple": {
			input: "@hmlt test() {\n= foo\n}",
			want: `Root
	Hmlt
		OutputCode
`,
		},
		"after element": {
			input: "@hmlt test() {\n%p= foo\n}",
			want: `Root
	Hmlt
		Element p()
			OutputCode
`,
		},
		"before content": {
			input: "@hmlt test() {\n= foo\n%p bar\n}",
			want: `Root
	Hmlt
		OutputCode
		Element p()
			Text(S)
`,
		},
		"mixed indents": {
			input: "@hmlt test() {\n%p1\n\t%p2= foo\n%p3 bar\n}",
			want: `Root
	Hmlt
		Element p1()
			Element p2()
				OutputCode
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
			input: "@hmlt test() {\n= @render foo()\n}",
			want: `Root
	Hmlt
		RenderCommand
`,
		},
		"with children": {
			input: "@hmlt test() {\n= @render foo()\n\t%p bar\n}",
			want: `Root
	Hmlt
		RenderCommand
			Element p()
				Text(S)
`,
		},
		"mixed indents": {
			input: "@hmlt test() {\n%p1 one\n= @render foo()\n\t%p bar\n%p2 two\n}",
			want: `Root
	Hmlt
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
			input: "@hmlt test() {\n= @children\n}",
			want: `Root
	Hmlt
		ChildrenCommand
`,
		},
		"mixed indents": {
			input: "@hmlt test() {\n%p1 one\n%parent\n\t= @children\n%p2 two\n}",
			want: `Root
	Hmlt
		Element p1()
			Text(S)
		Element parent()
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
