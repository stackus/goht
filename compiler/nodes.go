package compiler

import (
	"bytes"
	"fmt"
	"html"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/stackus/goht"
)

type nodeType int

const (
	nError nodeType = iota
	nRoot
	nGoCode
	nTemplate
	nIndent
	nDoctype
	nElement
	nNewLine
	nComment
	nText
	nRawText
	nUnescape
	nSilentScriptNode
	nScriptNode
	nRenderCommand
	nChildrenCommand
	nSlotCommand
	nFilter
)

func (n nodeType) String() string {
	switch n {
	case nError:
		return "Error"
	case nRoot:
		return "Root"
	case nGoCode:
		return "GoCode"
	case nTemplate:
		return "Template"
	case nIndent:
		return "Indent"
	case nDoctype:
		return "Doctype"
	case nElement:
		return "Element"
	case nNewLine:
		return "NewLine"
	case nComment:
		return "Comment"
	case nText:
		return "Text"
	case nRawText:
		return "RawText"
	case nUnescape:
		return "Unescape"
	case nSilentScriptNode:
		return "SilentScript"
	case nScriptNode:
		return "Script"
	case nRenderCommand:
		return "RenderCommand"
	case nChildrenCommand:
		return "ChildrenCommand"
	case nSlotCommand:
		return "SlotCommand"
	case nFilter:
		return "Filter"
	default:
		return "!Unknown!"
	}
}

type nodeBase interface {
	// Type returns the type of the node.
	Type() nodeType
	// Indent returns the indentation level of the node.
	Indent() int
	// Origin returns the origin of the node.
	Origin() token
	// Children return the children of the node.
	Children() []nodeBase
	// AddChild adds a child to the node.
	AddChild(nodeBase)
	// SetNextSibling sets the next sibling of the node.
	SetNextSibling(nodeBase)
	// Source returns the source code of the node.
	Source(tw *templateWriter) error
	Tree(buf *bytes.Buffer, indent int) string
}

type parsingNode interface {
	nodeBase
	// parse parses token data
	parse(*parser) error
}

type node struct {
	typ          nodeType
	indent       int
	origin       token
	children     []nodeBase
	nextSibling  nodeBase
	keepNewlines bool
}

func newNode(typ nodeType, indent int, origin token) node {
	return node{
		typ:    typ,
		indent: indent,
		origin: origin,
	}
}

func (n *node) Type() nodeType {
	return n.typ
}

func (n *node) Indent() int {
	return n.indent
}

func (n *node) Origin() token {
	return n.origin
}

func (n *node) Children() []nodeBase {
	return n.children
}

func (n *node) SetNextSibling(sibling nodeBase) {
	n.nextSibling = sibling
}

func (n *node) AddChild(c nodeBase) {
	if len(n.children) > 0 {
		n.children[len(n.children)-1].SetNextSibling(c)
	}
	n.children = append(n.children, c)
}

func (n *node) Tree(buf *bytes.Buffer, indent int) string {
	lead := strings.Repeat("\t", indent)
	buf.WriteString(lead + n.Type().String() + "\n")
	for _, c := range n.children {
		c.Tree(buf, indent+1)
	}
	return buf.String()
}

func (n *node) errorf(format string, args ...interface{}) error {
	return PositionalError{
		Line:   n.origin.line,
		Column: n.origin.col,
		Err:    fmt.Errorf(format, args...),
	}
}

func (n *node) handleNode(p *parser, indent int) error {
	t := p.peek()
	_ = t
	switch p.peek().Type() {
	case tKeepNewlines:
		n.keepNewlines = true
		p.next()
	case tRubyComment:
		p.next()
	case tNewLine:
		p.next()
		if n.keepNewlines {
			p.addChild(NewNewLineNode(t))
		}
	case tIndent:
		nextIndent := len(p.peek().lit)
		if nextIndent <= n.indent {
			return p.backToIndent(nextIndent - 1)
		}
		p.next()
		return n.handleNode(p, nextIndent)
	case tDoctype:
		p.addChild(NewDoctypeNode(p.next()))
	case tTag, tId, tClass:
		p.addNode(NewElementNode(p.next(), indent, n.keepNewlines))
	case tAttrName:
		p.addNode(NewElementNode(p.peek(), indent, n.keepNewlines))
	case tComment:
		p.addNode(NewCommentNode(p.next(), indent, n.keepNewlines))
	case tUnescaped:
		p.addNode(NewUnescapeNode(p.next(), indent))
	case tPlainText, tPreserveText, tEscapedText, tDynamicText:
		p.addChild(NewTextNode(p.next()))
	case tRawText:
		p.addNode(NewRawTextNode(p.next(), indent))
	case tSilentScript:
		p.addNode(NewSilentScriptNode(p.next(), indent, n.keepNewlines))
	case tScript:
		p.addChild(NewScriptNode(p.next(), n.keepNewlines))
	case tRenderCommand:
		p.addNode(NewRenderCommandNode(p.next(), indent, n.keepNewlines))
	case tChildrenCommand:
		p.addChild(NewChildrenCommandNode(p.next()))
	case tSlotCommand:
		p.addNode(NewSlotCommandNode(p.next(), indent, n.keepNewlines))
	case tFilterStart:
		t := p.next()
		switch t.lit {
		case "javascript":
			p.addNode(NewJavaScriptFilterNode(t, indent))
		case "css":
			p.addNode(NewCssFilterNode(t, indent))
		case "plain", "escaped", "preserve":
			p.addNode(NewTextFilterNode(t, indent))
		default:
			return n.errorf("unknown filter: %s", t)
		}
	case tTemplateEnd:
		return p.backToType(nTemplate)
	case tEOF:
		return n.errorf("template is incomplete: %s", p.peek())
	case tError:
		return PositionalError{
			Line:   t.line,
			Column: t.col,
			Err:    fmt.Errorf("%s", t.lit),
		}
	default:
		return n.errorf("unexpected: %s", p.peek())
	}
	return nil
}

type RootNode struct {
	node
	pkg         token
	imports     []string
	userImports []token
}

func NewRootNode() *RootNode {
	return &RootNode{
		node: newNode(nRoot, 0, token{typ: tRoot, line: 0, col: 0}),
		pkg:  token{typ: tPackage, line: 0, col: 0, lit: "main"},
		imports: []string{
			"\"context\"",
			"\"io\"",
			"\"github.com/stackus/goht\"",
		},
	}
}

func (n *RootNode) Source(tw *templateWriter) error {
	if _, err := tw.Write(
		fmt.Sprintf("// Code generated by GoHT %s - DO NOT EDIT.\n// https://github.com/stackus/goht\n\n",
			goht.Version(),
		)); err != nil {
		return err
	}

	if _, err := tw.Write("package "); err != nil {
		return err
	}
	r, err := tw.Write(n.pkg.lit)
	if err != nil {
		return err
	}
	tw.Add(n.pkg, r)
	if _, err := tw.Write("\n\n"); err != nil {
		return err
	}

	for _, pkg := range n.imports {
		if _, err := tw.Write("import " + pkg + "\n"); err != nil {
			return err
		}
	}

	if len(n.userImports) > 0 {
		if _, err := tw.Write("import (\n"); err != nil {
			return err
		}
		itw := tw.Indent(1)
		for _, pkg := range n.userImports {
			r, err := itw.WriteIndent(pkg.lit)
			if err != nil {
				return err
			}
			tw.Add(pkg, r)
			if _, err := itw.Write("\n"); err != nil {
				return err
			}
		}
		if _, err := tw.Write(")\n"); err != nil {
			return err
		}
	}

	for _, c := range n.children {
		if err := c.Source(tw); err != nil {
			return err
		}
	}
	return nil
}

func (n *RootNode) addImport(t token) {
	if slices.Contains(n.imports, t.lit) {
		return
	}
	for _, i := range n.userImports {
		if i.lit == t.lit {
			return
		}
	}
	n.userImports = append(n.userImports, t)
}

func (n *RootNode) parse(p *parser) error {
	switch p.peek().Type() {
	case tPackage:
		n.pkg = p.next()
	case tImport:
		t := p.next()
		n.addImport(t)
	case tGoCode, tNewLine:
		p.addNode(NewCodeNode(p.next()))
	case tTemplateStart:
		p.addNode(NewTemplateNode(p.next()))
	case tEOF:
		p.next()
		return nil
	default:
		return n.errorf("unexpected: %s", p.peek())
	}

	return nil
}

type CodeNode struct {
	node
	text   *strings.Builder
	tokens []token
}

func NewCodeNode(t token) *CodeNode {
	builder := &strings.Builder{}
	builder.WriteString(t.lit)
	return &CodeNode{
		node:   newNode(nGoCode, 0, t),
		text:   builder,
		tokens: []token{t},
	}
}

func (n *CodeNode) Source(tw *templateWriter) error {
	for _, t := range n.tokens {
		if r, err := tw.Write(t.lit); err != nil {
			return err
		} else if t.typ != tNewLine {
			tw.Add(t, r)
		}
	}
	return nil
}

func (n *CodeNode) parse(p *parser) error {
	switch p.peek().Type() {
	case tGoCode, tNewLine:
		t := p.next()
		_, err := n.text.WriteString(t.lit)
		if err != nil {
			return err
		}
		n.tokens = append(n.tokens, t)
		return nil
	case tPackage, tImport, tTemplateStart, tEOF:
		return p.backToType(nRoot)
	default:
		return n.errorf("unexpected: %s", p.peek())
	}
}

type TemplateNode struct {
	node
	decl string
}

func NewTemplateNode(t token) *TemplateNode {
	return &TemplateNode{
		node: newNode(nTemplate, -1, t),
		decl: t.lit,
	}
}

func (n *TemplateNode) Source(tw *templateWriter) error {
	entry := ` goht.Template {
	return goht.TemplateFunc(func(ctx context.Context, __w io.Writer, __sts ...goht.SlottedTemplate) (__err error) {
		__buf, __isBuf := __w.(goht.Buffer)
		if !__isBuf {
			__buf = goht.GetBuffer()
			defer goht.ReleaseBuffer(__buf)
		}
		var __children goht.Template
		ctx, __children = goht.PopChildren(ctx)
		_ = __children
`
	exit := `		if !__isBuf {
			_, __err = __w.Write(__buf.Bytes())
		}
		return
	})
}`
	tw.ResetVarName()
	if _, err := tw.Write("func "); err != nil {
		return err
	}
	if r, err := tw.Write(n.decl); err != nil {
		return err
	} else {
		tw.Add(n.origin, r)
	}
	if _, err := tw.Write(entry); err != nil {
		return err
	}

	itw := tw.Indent(2)
	for _, c := range n.children {
		if err := c.Source(itw); err != nil {
			return err
		}
	}
	// ensure the template ends with a newline
	if !n.keepNewlines {
		if _, err := itw.WriteStringLiteral("\\n"); err != nil {
			return err
		}
	}
	if _, err := itw.Close(); err != nil {
		return err
	}

	_, err := tw.Write(exit)
	return err
}

func (n *TemplateNode) parse(p *parser) error {
	switch p.peek().Type() {
	case tTemplateEnd:
		p.next()
		return p.backToType(nRoot)
	default:
		return n.handleNode(p, 0)
	}
}

type DoctypeNode struct {
	node
	doctype string
}

func NewDoctypeNode(t token) *DoctypeNode {
	return &DoctypeNode{
		node:    newNode(nDoctype, 0, t),
		doctype: t.lit,
	}
}

func (n *DoctypeNode) Source(tw *templateWriter) error {
	_, err := tw.WriteStringLiteral(`<!DOCTYPE html>`)
	return err
}

type attribute struct {
	name      string
	isBoolean bool
	isDynamic bool
	value     string
	origin    token
}

type ElementNode struct {
	node
	tag                 string
	id                  string
	classes             []token
	objectRef           *token
	attributes          *OrderedMap[attribute]
	attributesCmd       string
	disallowChildren    bool
	isSelfClosing       bool
	nukeInnerWhitespace bool
	nukeOuterWhitespace bool
	addWhitespaceBefore bool
	addWhitespaceAfter  bool
	isComplete          bool
}

func NewElementNode(t token, indent int, keepNewlines bool) *ElementNode {
	n := &ElementNode{
		node:       newNode(nElement, indent, t),
		tag:        "div",
		attributes: NewOrderedMap[attribute](),
	}

	switch t.Type() {
	case tTag:
		n.tag = t.lit
	case tId:
		n.id = t.lit
	case tClass:
		n.classes = append(n.classes, t)
	}

	if keepNewlines {
		n.keepNewlines = true
	}

	return n
}

func (n *ElementNode) Source(tw *templateWriter) error {
	if n.nukeOuterWhitespace {
		if _, err := tw.WriteStringLiteral(goht.NukeBefore); err != nil {
			return err
		}
	}
	if n.addWhitespaceBefore {
		if _, err := tw.WriteStringLiteral(" "); err != nil {
			return err
		}
	}

	if _, err := tw.WriteStringLiteral("<" + n.tag); err != nil {
		return err
	}
	// write attributes (and classes)
	if err := n.renderAttributes(tw); err != nil {
		return err
	}
	// close the tag
	if n.isSelfClosing {
		// add a "/" to the end of the tag as long as it's not in the list of tags that shouldn't get it
		if !slices.Contains(selfClosedTags, strings.ToLower(n.tag)) {
			if _, err := tw.WriteStringLiteral("/"); err != nil {
				return err
			}
		}
	}
	if _, err := tw.WriteStringLiteral(">"); err != nil {
		return err
	}
	if n.isSelfClosing {
		return nil
	}

	if n.nukeInnerWhitespace {
		if _, err := tw.WriteStringLiteral(goht.NukeAfter); err != nil {
			return err
		}
	}

	// ignore children if there is only a newline
	if !(len(n.children) == 1 && n.children[0].Type() == nNewLine) {
		for _, c := range n.children {
			if err := c.Source(tw); err != nil {
				return err
			}
		}
	}

	if n.nukeInnerWhitespace {
		if _, err := tw.WriteStringLiteral(goht.NukeBefore); err != nil {
			return err
		}
	}

	if _, err := tw.WriteStringLiteral("</" + n.tag + ">"); err != nil {
		return err
	}
	if n.nukeOuterWhitespace {
		if _, err := tw.WriteStringLiteral(goht.NukeAfter); err != nil {
			return err
		}
	} else if n.keepNewlines {
		if _, err := tw.WriteStringLiteral("\\n"); err != nil {
			return err
		}
	}
	if n.addWhitespaceAfter {
		if _, err := tw.WriteStringLiteral(" "); err != nil {
			return err
		}
	}

	return nil
}

func (n *ElementNode) renderAttributes(tw *templateWriter) error {
	if n.objectRef != nil {
		vName := tw.GetVarName()
		if _, err := tw.WriteIndent(`if ` + vName + ` := goht.ObjectID(`); err != nil {
			return err
		}
		if r, err := tw.Write(n.objectRef.lit); err != nil {
			return err
		} else {
			tw.Add(*n.objectRef, r)
		}
		if _, err := tw.Write("); " + vName + " != \"\" {\n"); err != nil {
			return err
		}
		if _, err := tw.WriteIndent("\t" + `if _, __err = __buf.WriteString(" id=\""+` + vName + `+"\""` + "); __err != nil { return }\n"); err != nil {
			return err
		}
		if _, err := tw.WriteIndent("}\n"); err != nil {
			return err
		}
	}
	if n.id != "" {
		if _, err := tw.WriteStringLiteral(` id=\"` + html.EscapeString(n.id) + `\"`); err != nil {
			return err
		}
	}
	if err := n.renderClass(tw); err != nil {
		return err
	}
	for _, key := range n.attributes.keys {
		attr := n.attributes.values[key]
		if attr.value == "" {
			if _, err := tw.WriteStringLiteral(` ` + attr.name); err != nil {
				return err
			}
			continue
		}
		if attr.isBoolean {
			if _, err := tw.WriteIndent("if "); err != nil {
				return err
			}
			if r, err := tw.Write(attr.value); err != nil {
				return err
			} else {
				tw.Add(attr.origin, r)
			}
			if _, err := tw.Write(" {\n"); err != nil {
				return err
			}
			itw := tw.Indent(1)
			if _, err := itw.WriteStringLiteral(" " + attr.name); err != nil {
				return err
			}
			if _, err := itw.Close(); err != nil {
				return err
			}
			if _, err := tw.WriteIndent("}\n"); err != nil {
				return err
			}
			continue
		}
		if _, err := tw.WriteStringLiteral(` ` + attr.name + `=\"`); err != nil {
			return err
		}
		if attr.isDynamic {
			if _, err := tw.WriteIndent(`if _, __err = __buf.WriteString(goht.EscapeString(`); err != nil {
				return err
			}
			if err := writeFormattedText(tw, attr.origin); err != nil {
				return err
			}
			if _, err := tw.Write(")+" + `"\""` + "); __err != nil { return }\n"); err != nil {
				return err
			}
			continue
		}
		if _, err := tw.WriteStringLiteral(html.EscapeString(attr.value) + `\"`); err != nil {
			return err
		}
	}
	if n.attributesCmd != "" {
		vName := tw.GetVarName()
		if _, err := tw.WriteIndent(`var ` + vName + " string\n"); err != nil {
			return err
		}
		if _, err := tw.WriteIndent(vName + `, __err = goht.BuildAttributeList(` + n.attributesCmd + ")\n"); err != nil {
			return err
		}
		if _, err := tw.WriteErrorHandler(); err != nil {
			return err
		}
		if _, err := tw.WriteStringIndent(`" "+` + vName); err != nil {
			return err
		}
	}
	return nil
}

func (n *ElementNode) renderClass(tw *templateWriter) error {
	if n.objectRef != nil {
		n.classes = append(n.classes, *n.objectRef)
	}
	if classes, ok := n.attributes.Get("class"); ok {
		if classes.isDynamic {
			n.classes = append(n.classes, classes.origin)
		} else {
			n.classes = append(n.classes, classes.origin)
		}
		n.attributes.Delete("class")
	}
	if len(n.classes) == 0 {
		return nil
	}
	allQuoted := true
	for _, class := range n.classes {
		switch class.typ {
		case tObjectRef, tAttrDynamicValue:
			allQuoted = false
		default:
		}
		if !allQuoted {
			break
		}
	}
	if allQuoted {
		classes := make([]string, len(n.classes))
		for i, class := range n.classes {
			var name string
			var err error
			switch class.typ {
			case tClass:
				name = class.lit
			case tAttrEscapedValue:
				name, err = strconv.Unquote(class.lit)
				if err != nil {
					return fmt.Errorf("failed to unquote class: %s error: %w", class.lit, err)
				}
			}
			classes[i] = name
		}
		_, err := tw.WriteStringLiteral(` class=\"` + strings.Join(classes, " ") + `\"`)
		return err
	}

	vName := tw.GetVarName()
	if _, err := tw.WriteIndent(`var ` + vName + " string\n"); err != nil {
		return err
	}
	if _, err := tw.WriteIndent(vName + ", __err = goht.BuildClassList("); err != nil {
		return err
	}
	for i, class := range n.classes {
		switch class.typ {
		case tObjectRef:
			if _, err := tw.Write("goht.ObjectClass(" + class.lit + ")"); err != nil {
				return err
			}
		case tAttrDynamicValue:
			if r, err := tw.Write(class.lit); err != nil {
				return err
			} else {
				tw.Add(class, r)
			}
		case tClass:
			if _, err := tw.Write(strconv.Quote(class.lit)); err != nil {
				return err
			}
		default:
			if _, err := tw.Write(class.lit); err != nil {
				return err
			}
		}
		if i < len(n.classes)-1 {
			if _, err := tw.Write(", "); err != nil {
				return err
			}
		}
	}
	if _, err := tw.Write(")\n"); err != nil {
		return err
	}
	if _, err := tw.WriteErrorHandler(); err != nil {
		return err
	}
	if _, err := tw.WriteStringIndent(`" class=\""+` + vName + `+"\""`); err != nil {
		return err
	}

	return nil
}

var selfClosedTags = []string{
	"area", "base", "basefont", "br", "col",
	"embed", "frame", "hr", "img", "input",
	"isindex", "keygen", "link", "menuitem",
	"meta", "param", "source", "track", "wbr",
}

func (n *ElementNode) parse(p *parser) error {
	t := p.peek()
	_ = t
	if n.isComplete {
		switch p.peek().Type() {
		case tIndent:
			nextIndent := len(p.peek().lit)
			if nextIndent <= n.indent {
				return p.backToIndent(nextIndent - 1)
			}
			if nextIndent > n.indent && (n.disallowChildren || n.isSelfClosing) {
				if n.isSelfClosing {
					return n.errorf("illegal nesting: self-closing tags can't have content %s", p.peek())
				}
				return n.errorf("illegal nesting: content can't be both given on the same line and nested %s", p.peek())
			}
		default:
			return n.handleNode(p, n.indent+1)
		}
	}
	switch p.peek().Type() {
	case tNewLine:
		p.next()
		n.isComplete = true
		if slices.Contains(selfClosedTags, n.tag) {
			n.isSelfClosing = true
		}
		if n.isSelfClosing || len(n.children) > 0 {
			n.disallowChildren = true
		}
		if len(n.children) == 0 && n.keepNewlines {
			n.AddChild(NewNewLineNode(t))
		}
	case tId:
		n.id = html.EscapeString(p.next().lit)
	case tClass:
		n.classes = append(n.classes, p.next())
	case tObjectRef:
		t := p.next()
		n.objectRef = &t
	case tAttrName:
		return n.parseAttributes(p)
	case tAttributesCommand:
		n.attributesCmd = p.next().lit
	case tVoidTag:
		p.next()
		n.isSelfClosing = true
	case tNukeOuterWhitespace:
		p.next()
		n.nukeOuterWhitespace = true
	case tNukeInnerWhitespace:
		p.next()
		n.nukeInnerWhitespace = true
	case tAddWhitespaceBefore:
		p.next()
		n.addWhitespaceBefore = true
	case tAddWhitespaceAfter:
		p.next()
		n.addWhitespaceAfter = true
	default:
		return n.handleNode(p, n.indent+1)
	}
	return nil
}

func (n *ElementNode) parseAttributes(p *parser) error {
	for {
		if p.peek().Type() != tAttrName {
			break
		}
		isBoolean := false
		isDynamic := false
		value := ""
		var origin token

		name := p.next().lit
		switch p.peek().Type() {
		case tAttrOperator:
			op := p.next().lit
			switch op {
			case "?":
				isBoolean = true
				if p.peek().Type() != tAttrDynamicValue {
					return n.errorf("expected dynamic value: %s", p.peek())
				}
			}
		default:
			n.attributes.Set(name, attribute{
				name: name,
			})
			continue
		}
		if p.peek().Type() != tAttrDynamicValue && p.peek().Type() != tAttrEscapedValue {
			return n.errorf("expected attribute value: %s", p.peek())
		}
		origin = p.next()
		if origin.typ == tAttrDynamicValue {
			isDynamic = true
			value = origin.lit
		} else {
			value, _ = strconv.Unquote(origin.lit)
		}

		n.attributes.Set(name, attribute{
			name:      name,
			isBoolean: isBoolean,
			isDynamic: isDynamic,
			value:     value,
			origin:    origin,
		})
	}
	return nil
}

func (n *ElementNode) Tree(buf *bytes.Buffer, indent int) string {
	lead := strings.Repeat("\t", indent)
	// build a list of attributes
	attrs := make([]string, 0, n.attributes.Len())

	err := n.attributes.Range(func(_ string, attr attribute) (bool, error) {
		a := attr.name
		if attr.value == "" {
			attrs = append(attrs, a)
			return true, nil
		}
		if attr.isBoolean {
			a += "?"
		}
		a += "="
		if attr.isDynamic {
			a += "{" + attr.value + "}"
		} else {
			a += fmt.Sprintf("%q", attr.value)
		}
		attrs = append(attrs, a)
		return true, nil
	})
	if err != nil {
		panic(err)
	}
	if n.attributesCmd != "" {
		attrs = append(attrs, "@attrs={"+n.attributesCmd+"...}")
	}

	buf.WriteString(lead + n.Type().String() + " " + n.tag + "(" + strings.Join(attrs, ",") + ")\n")
	for _, c := range n.children {
		c.Tree(buf, indent+1)
	}
	return buf.String()
}

type NewLineNode struct {
	node
}

func NewNewLineNode(t token) *NewLineNode {
	return &NewLineNode{
		node: newNode(nNewLine, 0, t),
	}
}

func (n *NewLineNode) Source(tw *templateWriter) error {
	_, err := tw.WriteStringLiteral("\\n")
	return err
}

type CommentNode struct {
	node
	text string
}

func NewCommentNode(t token, indent int, keepNewlines bool) *CommentNode {
	n := &CommentNode{
		node: newNode(nComment, indent, t),
		text: t.lit,
	}

	if keepNewlines {
		n.keepNewlines = true
	}

	return n
}

func (n *CommentNode) Source(tw *templateWriter) error {
	if n.text != "" {
		if _, err := tw.WriteStringLiteral("<!--" + html.EscapeString(n.text) + "-->"); err != nil {
			return err
		}
		if n.keepNewlines {
			if _, err := tw.WriteStringLiteral("\\n"); err != nil {
				return err
			}
		}
	}
	// ignore children if there is only a newline
	if len(n.children) == 0 || len(n.children) == 1 && n.children[0].Type() == nNewLine {
		return nil
	}
	if _, err := tw.WriteStringLiteral("<!--"); err != nil {
		return err
	}

	for _, c := range n.children {
		if err := c.Source(tw); err != nil {
			return err
		}
	}

	if _, err := tw.WriteStringLiteral("-->"); err != nil {
		return err
	}
	if n.keepNewlines {
		if _, err := tw.WriteStringLiteral("\\n"); err != nil {
			return err
		}
	}

	return nil
}

func (n *CommentNode) parse(p *parser) error {
	if p.peek().Type() == tIndent {
		nextIndent := len(p.peek().lit)
		if nextIndent <= n.indent {
			return p.backToIndent(nextIndent - 1)
		}
		if nextIndent > n.indent && n.text != "" {
			return n.errorf("illegal nesting: content can't be both given on the same line and nested %s", p.peek())
		}
	}
	return n.handleNode(p, n.indent+1)
}

type TextNode struct {
	node
	text       string
	isDynamic  bool
	isPlain    bool
	isPreserve bool
}

func NewTextNode(t token) *TextNode {
	return &TextNode{
		node:       newNode(nText, 0, t),
		text:       t.lit,
		isDynamic:  t.Type() == tDynamicText,
		isPlain:    t.Type() == tPlainText,
		isPreserve: t.Type() == tPreserveText,
	}
}

var reFmtText = regexp.MustCompile(`\A(%[^ ]+?) (.+)\z`)

func writeFormattedText(tw *templateWriter, t token) error {
	if matches := reFmtText.FindStringSubmatch(t.lit); matches != nil {
		if _, err := tw.Write(`goht.FormatString("`); err != nil {
			return err
		}
		if r, err := tw.Write(matches[1]); err != nil {
			return err
		} else {
			fmtT := token{
				typ:  tDynamicText,
				line: t.line,
				col:  t.col,
				lit:  matches[1],
			}
			tw.Add(fmtT, r)
		}
		if _, err := tw.Write(`", `); err != nil {
			return err
		}
		if r, err := tw.Write(matches[2]); err != nil {
			return err
		} else {
			strT := token{
				typ:  tDynamicText,
				line: t.line,
				col:  t.col + len(t.lit) - len(matches[2]),
				lit:  matches[2],
			}
			tw.Add(strT, r)
		}
		if _, err := tw.Write(")"); err != nil {
			return err
		}
		return nil
	}
	if r, err := tw.Write(strings.TrimSpace(strings.Replace(t.lit, "\n", " ", -1))); err != nil {
		return err
	} else {
		tw.Add(t, r)
	}
	return nil
}

func (n *TextNode) Source(tw *templateWriter) error {
	if n.isDynamic {
		vName := tw.GetVarName()
		if _, err := tw.WriteIndent(`var ` + vName + " string\n"); err != nil {
			return err
		}

		if _, err := tw.WriteIndent(`if ` + vName + `, __err = goht.CaptureErrors(`); err != nil {
			return err
		}
		if !tw.isUnescaped {
			if _, err := tw.Write(`goht.EscapeString(`); err != nil {
				return err
			}
		}
		if err := writeFormattedText(tw, n.origin); err != nil {
			return err
		}
		if !tw.isUnescaped {
			if _, err := tw.Write(")"); err != nil {
				return err
			}
		}
		if _, err := tw.Write("); __err != nil { return }\n"); err != nil {
			return err
		}

		if _, err := tw.WriteStringIndent(vName); err != nil {
			return err
		}
		return nil
	}

	s := n.text
	s = strconv.Quote(n.text)
	s = s[1 : len(s)-1]

	if n.isPreserve {
		start := len(s)
		s = strings.TrimSuffix(s, "\\n")
		s += strings.Repeat("&#x000A;", (start-len(s))/2)
	}

	if n.isPlain || n.isPreserve || tw.isUnescaped {
		_, err := tw.WriteStringLiteral(s)
		return err
	}

	_, err := tw.WriteStringLiteral(html.EscapeString(s))
	return err
}

func (n *TextNode) Tree(buf *bytes.Buffer, indent int) string {
	lead := strings.Repeat("\t", indent)
	typ := "(S)"
	if n.isDynamic {
		typ = "(D)"
	}
	buf.WriteString(lead + n.Type().String() + typ + "\n")
	for _, c := range n.children {
		c.Tree(buf, indent+1)
	}
	return buf.String()
}

type RawTextNode struct {
	node
	text string
}

func NewRawTextNode(t token, indent int) *RawTextNode {
	return &RawTextNode{
		node: newNode(nText, indent, t),
		text: t.lit,
	}
}

func (n *RawTextNode) Source(tw *templateWriter) error {
	if n.text != "" {
		s := n.text
		s = strconv.Quote(n.text)
		s = s[1 : len(s)-1]
		if _, err := tw.WriteStringLiteral(s); err != nil {
			return err
		}
	}

	return nil
}

func (n *RawTextNode) Tree(buf *bytes.Buffer, indent int) string {
	lead := strings.Repeat("\t", indent)
	buf.WriteString(lead + n.Type().String() + "\n")
	return buf.String()
}

func (n *RawTextNode) parse(p *parser) error {
	return p.backToParent()
}

type UnescapeNode struct {
	node
}

func NewUnescapeNode(t token, indent int) *UnescapeNode {
	return &UnescapeNode{
		node: newNode(nUnescape, indent, t),
	}
}

func (n *UnescapeNode) Source(tw *templateWriter) error {
	tw.isUnescaped = true
	for _, child := range n.children {
		err := child.Source(tw)
		if err != nil {
			return err
		}
	}
	tw.isUnescaped = false
	return nil
}

func (n *UnescapeNode) parse(p *parser) error {
	t := p.peek()
	_ = t
	switch p.peek().Type() {
	case tNewLine:
		return p.backToParent()
	default:
		return n.handleNode(p, n.indent)
	}
}

type SilentScriptNode struct {
	node
	code       string
	needsClose bool
	forRender  bool
}

func NewSilentScriptNode(t token, indent int, keepNewlines bool) *SilentScriptNode {
	n := &SilentScriptNode{
		node: newNode(nSilentScriptNode, indent, t),
		code: t.lit,
	}

	if keepNewlines {
		n.keepNewlines = true
	}

	return n
}

var openingStatements = []string{"if", "else if", "for", "switch"}
var elseStatements = []string{"else", "else if"}

func (n *SilentScriptNode) allowChildren() bool {
	code := strings.TrimSpace(strings.Replace(n.code, "\n", " ", -1))

	allowChildren := false
	for _, statement := range openingStatements {
		if strings.HasPrefix(code, statement) {
			allowChildren = true
			break
		}
	}
	if !allowChildren {
		for _, statement := range elseStatements {
			if strings.HasPrefix(code, statement) {
				allowChildren = true
				break
			}
		}
	}
	if strings.HasSuffix(code, "{") {
		allowChildren = true
	}

	return allowChildren
}

func (n *SilentScriptNode) Source(tw *templateWriter) error {
	if n.forRender {
		return nil
	}

	code := strings.TrimSpace(strings.Replace(n.code, "\n", " ", -1))

	isOpening := false
	for _, statement := range openingStatements {
		if strings.HasPrefix(code, statement) {
			isOpening = true
			break
		}
	}

	start := ""
	end := "\n"
	if n.needsClose && !strings.HasPrefix(code, "}") {
		start = "} "
	}
	if len(n.children) > 0 {
		if isOpening && !strings.HasSuffix(code, "{") {
			end = " {\n"
		}
	}

	if _, err := tw.WriteIndent(start); err != nil {
		return err
	}
	if r, err := tw.Write(code); err != nil {
		return err
	} else {
		tw.Add(n.origin, r)
	}
	if _, err := tw.Write(end); err != nil {
		return err
	}

	if len(n.children) == 0 {
		return nil
	}

	itw := tw.Indent(1)
	for _, c := range n.children {
		if err := c.Source(itw); err != nil {
			return err
		}
	}

	if _, err := itw.Close(); err != nil {
		return err
	}

	// only do any of this when n is the opening statement
	if isOpening {
		if next, ok := n.nextSibling.(*SilentScriptNode); ok {
			if strings.HasPrefix(strings.TrimSpace(next.code), "}") {
				return nil
			}
			for _, stmt := range elseStatements {
				if strings.HasPrefix(next.code, stmt) {
					next.needsClose = true
					return nil
				}
			}
		}
		// either there's no next SilentScript, or it's not an else-type, so close now
		if _, err := tw.WriteIndent("}\n"); err != nil {
			return err
		}
	}

	return nil
}

func (n *SilentScriptNode) parse(p *parser) error {
	switch p.peek().Type() {
	case tNewLine:
		p.next()
		return nil
	default:
		if !n.allowChildren() {
			return p.backToParent()
		}
		return n.handleNode(p, n.indent+1)
	}
}

type ScriptNode struct {
	node
	code string
}

func NewScriptNode(t token, keepNewlines bool) *ScriptNode {
	n := &ScriptNode{
		node: newNode(nScriptNode, 0, t),
		code: t.lit,
	}

	if keepNewlines {
		n.keepNewlines = true
	}

	return n
}

func (n *ScriptNode) Source(tw *templateWriter) error {
	vName := tw.GetVarName()
	if _, err := tw.WriteIndent(`var ` + vName + " string\n"); err != nil {
		return err
	}
	if _, err := tw.WriteIndent(`if ` + vName + `, __err = goht.CaptureErrors(`); err != nil {
		return err
	}
	if !tw.isUnescaped {
		if _, err := tw.Write(`goht.EscapeString(`); err != nil {
			return err
		}
	}
	if err := writeFormattedText(tw, n.origin); err != nil {
		return err
	}
	if !tw.isUnescaped {
		if _, err := tw.Write(")"); err != nil {
			return err
		}
	}
	if _, err := tw.Write("); __err != nil { return }\n"); err != nil {
		return err
	}
	if _, err := tw.WriteStringIndent(vName); err != nil {
		return err
	}
	return nil
}

type RenderCommandNode struct {
	node
	command string
}

func NewRenderCommandNode(t token, indent int, keepNewlines bool) *RenderCommandNode {
	n := &RenderCommandNode{
		node:    newNode(nRenderCommand, indent, t),
		command: t.lit,
	}

	if keepNewlines {
		n.keepNewlines = true
	}

	return n
}

func (n *RenderCommandNode) Source(tw *templateWriter) error {
	if len(n.children) == 0 {
		if _, err := tw.WriteIndent("if __err = "); err != nil {
			return err
		}
		if r, err := tw.Write(strings.TrimSpace(strings.Replace(n.command, "\n", " ", -1))); err != nil {
			return err
		} else {
			tw.Add(n.origin, r)
		}
		if _, err := tw.Write(".Render(ctx, __buf); __err != nil { return }\n"); err != nil {
			return err
		}
		return nil
	}

	vName := tw.GetVarName()

	fnLine := vName + " := goht.TemplateFunc(func(ctx context.Context, __w io.Writer, _ ...goht.SlottedTemplate) (__err error) {\n"

	if _, err := tw.WriteIndent(fnLine); err != nil {
		return err
	}

	itw := tw.Indent(1)

	lines := []string{
		"__buf, __isBuf := __w.(goht.Buffer)\n",
		"if !__isBuf {\n",
		"	__buf = goht.GetBuffer()\n",
		"	defer goht.ReleaseBuffer(__buf)\n",
		"}\n",
	}
	for _, line := range lines {
		if _, err := itw.WriteIndent(line); err != nil {
			return err
		}
	}
	for _, c := range n.children {
		if err := c.Source(itw); err != nil {
			return err
		}
	}
	if _, err := itw.Close(); err != nil {
		return err
	}

	lines = []string{
		"	if !__isBuf {\n",
		"		_, __err = io.Copy(__w, __buf)\n",
		"	}\n",
		"	return\n",
		"})\n",
	}
	for _, line := range lines {
		if _, err := tw.WriteIndent(line); err != nil {
			return err
		}
	}

	if _, err := tw.WriteIndent("if __err = "); err != nil {
		return err
	}
	if r, err := tw.Write(strings.TrimRight(n.command, " \t{")); err != nil {
		return err
	} else {
		tw.Add(n.origin, r)
	}
	if _, err := tw.Write(".Render(goht.PushChildren(ctx, " + vName + "), __buf, __sts...); __err != nil { return }\n"); err != nil {
		return err
	}

	if p, ok := n.nextSibling.(*SilentScriptNode); ok {
		// skip the next sibling if it is a closing brace
		if strings.TrimSpace(p.code) == "}" {
			p.forRender = true
		}
	}
	return nil
}

func (n *RenderCommandNode) parse(p *parser) error {
	switch p.peek().Type() {
	case tNewLine:
		p.next()
		return nil
	default:
		return n.handleNode(p, n.indent+1)
	}
}

type ChildrenCommandNode struct {
	node
}

func NewChildrenCommandNode(t token) *ChildrenCommandNode {
	return &ChildrenCommandNode{
		node: newNode(nChildrenCommand, 0, t),
	}
}

func (n *ChildrenCommandNode) Source(tw *templateWriter) error {
	_, err := tw.WriteIndent("if __err = __children.Render(ctx, __buf, __sts...); __err != nil { return }\n")
	return err
}

type SlotCommandNode struct {
	node
	slot string
}

func NewSlotCommandNode(t token, indent int, keepNewlines bool) *SlotCommandNode {
	n := &SlotCommandNode{
		node: newNode(nSlotCommand, indent, t),
		slot: t.lit,
	}

	if keepNewlines {
		n.keepNewlines = true
	}

	return n
}

func (n *SlotCommandNode) Source(tw *templateWriter) error {
	if _, err := tw.WriteIndent("if __st := goht.GetSlottedTemplate(__sts, " + strconv.Quote(n.slot) + "); __st != nil {\n"); err != nil {
		return err
	}

	itw := tw.Indent(1)

	// lines := []string{
	// 			"__sts := append(__st.SlottedTemplates(), __sts...)\n",
	// 			"_ = __sts\n",
	// }
	// for _, line := range lines {
	// 	if _, err := itw.WriteIndent(line); err != nil {
	// 		return err
	// 	}
	// }

	if _, err := itw.WriteIndent("if __err = __st.Render(ctx, __buf, append(__st.SlottedTemplates(), __sts...)...); __err != nil { return }\n"); err != nil {
		return err
	}

	if _, err := itw.Close(); err != nil {
		return err
	}

	if len(n.children) == 0 {
		_, err := tw.WriteIndent("}\n")
		return err
	}

	if _, err := tw.WriteIndent("} else {\n"); err != nil {
		return err
	}

	itw = tw.Indent(1)
	for _, c := range n.children {
		if err := c.Source(itw); err != nil {
			return err
		}
	}

	if _, err := itw.Close(); err != nil {
		return err
	}

	if next, ok := n.nextSibling.(*SilentScriptNode); ok {
		if strings.TrimSpace(next.code) == "}" {
			return nil
		}
	}
	// either there's no next SilentScript, or it's not an else-type, so close now
	if _, err := tw.WriteIndent("}\n"); err != nil {
		return err
	}

	return nil
}

func (n *SlotCommandNode) parse(p *parser) error {
	switch p.peek().Type() {
	case tNewLine:
		p.next()
		return nil
	default:
		return n.handleNode(p, n.indent+1)
	}
}

type JavaScriptFilterNode struct {
	node
}

func NewJavaScriptFilterNode(t token, indent int) *JavaScriptFilterNode {
	return &JavaScriptFilterNode{
		node: newNode(nFilter, indent, t),
	}
}

func (n *JavaScriptFilterNode) Source(tw *templateWriter) error {
	if _, err := tw.WriteStringLiteral("<script>\\n"); err != nil {
		return err
	}
	for _, c := range n.children {
		if err := c.Source(tw); err != nil {
			return err
		}
	}
	_, err := tw.WriteStringLiteral("</script>")
	return err
}

func (n *JavaScriptFilterNode) parse(p *parser) error {
	switch p.peek().Type() {
	case tPlainText, tDynamicText:
		n.AddChild(NewTextNode(p.next()))
	case tFilterEnd:
		p.next()
		return p.backToParent()
	case tEOF:
		return n.errorf("javascript filter is incomplete: %s", p.peek())
	default:
		return n.errorf("unexpected token: %s", p.peek())
	}
	return nil
}

type CssFilterNode struct {
	node
}

func NewCssFilterNode(t token, indent int) *CssFilterNode {
	return &CssFilterNode{
		node: newNode(nFilter, indent, t),
	}
}

func (n *CssFilterNode) Source(tw *templateWriter) error {
	if _, err := tw.WriteStringLiteral("<style>\\n"); err != nil {
		return err
	}
	for _, c := range n.children {
		if err := c.Source(tw); err != nil {
			return err
		}
	}
	_, err := tw.WriteStringLiteral("</style>")
	return err
}

func (n *CssFilterNode) parse(p *parser) error {
	switch p.peek().Type() {
	case tPlainText, tDynamicText:
		n.AddChild(NewTextNode(p.next()))
	case tFilterEnd:
		p.next()
		return p.backToParent()
	case tEOF:
		return n.errorf("css filter is incomplete: %s", p.peek())
	default:
		return n.errorf("unexpected token: %s", p.peek())
	}
	return nil
}

type TextFilterNode struct {
	node
	isUnescaped bool
}

func NewTextFilterNode(t token, indent int) *TextFilterNode {
	return &TextFilterNode{
		node:        newNode(nFilter, indent, t),
		isUnescaped: t.lit == "plain" || t.lit == "preserve",
	}
}

func (n *TextFilterNode) Source(tw *templateWriter) error {
	if n.isUnescaped {
		tw.isUnescaped = true
		defer func() {
			tw.isUnescaped = false
		}()
	}
	for _, c := range n.children {
		if err := c.Source(tw); err != nil {
			return err
		}
	}
	if n.origin.lit == "preserve" {
		_, err := tw.WriteStringLiteral("\\n")
		return err
	}

	return nil
}

func (n *TextFilterNode) parse(p *parser) error {
	switch p.peek().Type() {
	case tPlainText, tEscapedText, tPreserveText, tDynamicText:
		n.AddChild(NewTextNode(p.next()))
	case tFilterEnd:
		p.next()
		return p.backToParent()
	case tEOF:
		return n.errorf("text filter is incomplete: %s", p.peek())
	default:
		return n.errorf("unexpected token: %s", p.peek())
	}
	return nil
}
