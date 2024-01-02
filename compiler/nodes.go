package compiler

import (
	"bytes"
	"fmt"
	"html"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/stackus/hamlet"
)

type nodeType int

const (
	nError nodeType = iota
	nRoot
	nGoCode
	nHmlt
	nIndent
	nDoctype
	nElement
	nNewLine
	nComment
	nText
	nUnescape
	nSilentScriptNode
	nScriptNode
	nRenderCommand
	nChildrenCommand
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
	case nHmlt:
		return "Hmlt"
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
	typ      nodeType
	indent   int
	origin   token
	children []nodeBase
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

func (n *node) AddChild(c nodeBase) {
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
	return fmt.Errorf("%s[%d,%d]: %s", n.origin.typ, n.origin.line, n.origin.col, fmt.Sprintf(format, args...))
}

func (n *node) handleNode(p *parser, indent int) error {
	t := p.peek()
	_ = t
	switch p.peek().Type() {
	case tRubyComment:
		p.next()
	case tNewLine:
		// p.next()
		p.addChild(NewNewLineNode(p.next()))
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
		p.addNode(NewElementNode(p.next(), indent))
	case tComment:
		p.addNode(NewCommentNode(p.next(), indent))
	case tUnescaped:
		p.addNode(NewUnescapeNode(p.next(), indent))
	case tPlainText, tPreserveText, tEscapedText, tDynamicText:
		p.addChild(NewTextNode(p.next()))
	case tSilentScript:
		p.addNode(NewSilentScriptNode(p.next(), indent))
	case tScript:
		p.addChild(NewScriptNode(p.next()))
	case tRenderCommand:
		p.addNode(NewRenderCommandNode(p.next(), indent))
	case tChildrenCommand:
		p.addChild(NewChildrenCommandNode(p.next()))
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
	case tHmltEnd:
		return p.backToType(nHmlt)
	case tEOF:
		return n.errorf("template is incomplete: %s", p.peek())
	default:
		return n.errorf("unexpected: %s", p.peek())
	}
	return nil
}

type RootNode struct {
	node
	pkg     string
	imports []string
}

func NewRootNode() *RootNode {
	return &RootNode{
		node: newNode(nRoot, 0, token{typ: tRoot, line: 0, col: 0}),
		pkg:  "main",
		imports: []string{
			"\"bytes\"",
			"\"context\"",
			"\"io\"",
			"\"github.com/stackus/hamlet\"",
		},
	}
}

func (n *RootNode) Source(tw *templateWriter) error {
	if err := tw.Write("package " + n.pkg + "\n\n"); err != nil {
		return err
	}

	if err := tw.Write("import (\n"); err != nil {
		return err
	}
	itw := tw.Indent()
	for _, pkg := range n.imports {
		if err := itw.Write(pkg + "\n"); err != nil {
			return err
		}
	}
	if err := tw.Write(")\n"); err != nil {
		return err
	}

	for _, c := range n.children {
		if err := c.Source(tw); err != nil {
			return err
		}
	}
	return nil
}

func (n *RootNode) parse(p *parser) error {
	switch p.peek().Type() {
	case tPackage:
		n.pkg = p.next().lit
	case tImport:
		t := p.next()
		if !slices.Contains(n.imports, t.lit) {
			n.imports = append(n.imports, t.lit)
		}
	case tGoCode, tNewLine:
		p.addNode(NewCodeNode(p.next()))
	case tHmltStart:
		p.addNode(NewHmltNode(p.next()))
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
	text *strings.Builder
}

func NewCodeNode(t token) *CodeNode {
	builder := &strings.Builder{}
	builder.WriteString(t.lit)
	return &CodeNode{
		node: newNode(nGoCode, 0, t),
		text: builder,
	}
}

func (n *CodeNode) Source(tw *templateWriter) error {
	return tw.Write(n.text.String())
}

func (n *CodeNode) parse(p *parser) error {
	switch p.peek().Type() {
	case tGoCode, tNewLine:
		_, err := n.text.WriteString(p.next().lit)
		return err
	case tImport, tHmltStart, tEOF:
		return p.backToType(nRoot)
	default:
		return n.errorf("unexpected: %s", p.peek())
	}
}

type HmltNode struct {
	node
	decl string
}

func NewHmltNode(t token) *HmltNode {
	return &HmltNode{
		node: newNode(nHmlt, -1, t),
		decl: t.lit,
	}
}

func (n *HmltNode) Source(tw *templateWriter) error {
	entry := `func %s hamlet.Template {
	return hamlet.TemplateFunc(func(ctx context.Context, __w io.Writer) (__err error) {
		__buf, __isBuf := __w.(*bytes.Buffer)
		if !__isBuf {
			__buf = hamlet.GetBuffer()
			defer hamlet.ReleaseBuffer(__buf)
		}
		var __children hamlet.Template
		ctx, __children = hamlet.PopChildren(ctx)
		_ = __children
`
	exit := `		if !__isBuf {
			_, __err = __w.Write(hamlet.NukeWhitespace(__buf.Bytes()))
		}
		return
	})
}
`
	if err := tw.Write(fmt.Sprintf(entry, n.decl)); err != nil {
		return err
	}
	for _, c := range n.children {

		if err := c.Source(tw); err != nil {
			return err
		}
	}
	return tw.Write(exit)
}

func (n *HmltNode) parse(p *parser) error {
	switch p.peek().Type() {
	case tHmltEnd:
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
	return tw.WriteStringLiteral(`<!DOCTYPE html>`)
}

type attribute struct {
	name      string
	isBoolean bool
	isDynamic bool
	value     string
}

type ElementNode struct {
	node
	tag                 string
	id                  string
	classes             []string
	objectRef           string
	attributes          *OrderedMap[attribute]
	attributesCmd       string
	disallowChildren    bool
	isSelfClosing       bool
	nukeInnerWhitespace bool
	nukeOuterWhitespace bool
	isComplete          bool
}

func NewElementNode(t token, indent int) *ElementNode {
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
		n.classes = append(n.classes, fmt.Sprintf("%q", t.lit))
	}

	return n
}

func (n *ElementNode) Source(tw *templateWriter) error {
	if n.nukeOuterWhitespace {
		if err := tw.WriteStringLiteral(hamlet.NukeBefore); err != nil {
			return err
		}
	}

	if err := tw.WriteStringLiteral("<" + n.tag); err != nil {
		return err
	}
	// write attributes (and classes)
	if err := n.renderAttributes(tw); err != nil {
		return err
	}
	// close the tag
	if err := tw.WriteStringLiteral(">"); err != nil {
		return err
	}
	if n.isSelfClosing {
		return nil
	}

	if n.nukeInnerWhitespace {
		if err := tw.WriteStringLiteral(hamlet.NukeAfter); err != nil {
			return err
		}
	}

	for _, c := range n.children {
		if err := c.Source(tw); err != nil {
			return err
		}
	}

	if n.nukeInnerWhitespace {
		if err := tw.WriteStringLiteral(hamlet.NukeBefore); err != nil {
			return err
		}
	}

	if err := tw.WriteStringLiteral("</" + n.tag + ">"); err != nil {
		return err
	}

	if n.nukeOuterWhitespace {
		if err := tw.WriteStringLiteral(hamlet.NukeAfter); err != nil {
			return err
		}
	} else {
		if err := tw.WriteStringLiteral("\\n"); err != nil {
			return err
		}
	}

	return nil
}

func (n *ElementNode) renderAttributes(tw *templateWriter) error {
	if n.objectRef != "" {
		vName := tw.GetVarName()
		if err := tw.WriteIndent(`if ` + vName + ` := hamlet.ObjectID(` + n.objectRef + `); ` + vName + " != \"\" {\n"); err != nil {
			return err
		}
		if err := tw.WriteIndent("\t" + `if _, __err = __buf.WriteString(" id=\""+` + vName + `+"\""` + "); __err != nil { return }\n"); err != nil {
			return err
		}
		if err := tw.WriteIndent("}\n"); err != nil {
			return err
		}
	}
	if n.id != "" {
		if err := tw.WriteStringLiteral(` id=\"` + html.EscapeString(n.id) + `\"`); err != nil {
			return err
		}
	}
	if err := n.renderClass(tw); err != nil {
		return err
	}
	for _, key := range n.attributes.keys {
		attr := n.attributes.values[key]
		if attr.value == "" {
			if err := tw.WriteStringLiteral(` ` + attr.name); err != nil {
				return err
			}
			continue
		}
		if attr.isBoolean {
			if err := tw.WriteIndent(`if ` + attr.value + " {\n"); err != nil {
				return err
			}
			itw := tw.Indent()
			if err := itw.WriteStringLiteral(" " + attr.name); err != nil {
				return err
			}
			if err := itw.Close(); err != nil {
				return err
			}
			if err := tw.WriteIndent("}\n"); err != nil {
				return err
			}
			continue
		}
		if err := tw.WriteStringLiteral(` ` + attr.name + `=\"`); err != nil {
			return err
		}
		if attr.isDynamic {
			value := attr.value
			if matches := reFmtText.FindStringSubmatch(attr.value); matches != nil {
				value = `hamlet.FormatString("` + matches[1] + `", ` + matches[2] + `)`
			}
			if err := tw.WriteStringIndent(`hamlet.EscapeString(` + value + `)+"\""`); err != nil {
				return err
			}
			continue
		}
		if err := tw.WriteStringLiteral(html.EscapeString(attr.value) + `\"`); err != nil {
			return err
		}
	}
	if n.attributesCmd != "" {
		vName := tw.GetVarName()
		if err := tw.WriteIndent(`var ` + vName + " string\n"); err != nil {
			return err
		}
		if err := tw.WriteIndent(vName + `, __err = hamlet.BuildAttributeList(` + n.attributesCmd + ")\n"); err != nil {
			return err
		}
		if err := tw.WriteErrorHandler(); err != nil {
			return err
		}
		if err := tw.WriteStringIndent(vName); err != nil {
			return err
		}
	}
	return nil
}

func (n *ElementNode) renderClass(tw *templateWriter) error {
	if n.objectRef != "" {
		n.classes = append(n.classes, "hamlet.ObjectClass("+n.objectRef+")")
	}
	if classes, ok := n.attributes.Get("class"); ok {
		if classes.isDynamic {
			n.classes = append(n.classes, classes.value)
		} else {
			n.classes = append(n.classes, strconv.Quote(html.EscapeString(classes.value)))
		}
		n.attributes.Delete("class")
	}
	if len(n.classes) == 0 {
		return nil
	}
	allQuoted := true
	for _, class := range n.classes {
		if !strings.HasPrefix(class, "\"") {
			allQuoted = false
			break
		}
	}
	if allQuoted {
		classList := strings.Builder{}
		for _, class := range n.classes {
			classList.WriteString(class[1 : len(class)-1])
			classList.WriteString(" ")
		}
		return tw.WriteStringLiteral(` class=\"` + strings.TrimSpace(classList.String()) + `\"`)
	}

	if len(n.classes) > 0 {
		classList := strings.Join(n.classes, ", ")
		vName := tw.GetVarName()
		if err := tw.WriteIndent(`var ` + vName + " string\n"); err != nil {
			return err
		}
		if err := tw.WriteIndent(vName + `, __err = hamlet.BuildClassList(` + classList + ")\n"); err != nil {
			return err
		}
		if err := tw.WriteErrorHandler(); err != nil {
			return err
		}
		if err := tw.WriteStringIndent(`" class=\""+` + vName + `+"\""`); err != nil {
			return err
		}
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
		t := p.next()
		n.isComplete = true
		if slices.Contains(selfClosedTags, n.tag) {
			n.isSelfClosing = true
		}
		if n.isSelfClosing || len(n.children) > 0 {
			n.disallowChildren = true
		}
		if len(n.children) == 0 {
			n.AddChild(NewNewLineNode(t))
		}
	case tId:
		n.id = html.EscapeString(p.next().lit)
	case tClass:
		n.classes = append(n.classes, strconv.Quote(html.EscapeString(p.next().lit)))
	case tObjectRef:
		n.objectRef = p.next().lit
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

		name := p.next().lit
		switch p.peek().Type() {
		case tAttrName:
			n.attributes.Set(name, attribute{
				name: name,
			})
			continue
		case tAttrOperator:
			op := p.next().lit
			switch op {
			case "?":
				isBoolean = true
				if p.peek().Type() != tAttrDynamicValue {
					return n.errorf("expected dynamic value: %s", p.peek())
				}
			}
		}
		if p.peek().Type() != tAttrDynamicValue && p.peek().Type() != tAttrEscapedValue {
			return n.errorf("expected attribute value: %s", p.peek())
		}
		if p.peek().Type() == tAttrDynamicValue {
			isDynamic = true
			value = p.next().lit
		} else {
			value, _ = strconv.Unquote(p.next().lit)
		}

		n.attributes.Set(name, attribute{
			name:      name,
			isBoolean: isBoolean,
			isDynamic: isDynamic,
			value:     value,
		})
	}
	return nil
}

func (n *ElementNode) Tree(buf *bytes.Buffer, indent int) string {
	lead := strings.Repeat("\t", indent)
	// build list of attributes
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
	return tw.WriteStringLiteral("\\n")
}

type CommentNode struct {
	node
	text string
}

func NewCommentNode(t token, indent int) *CommentNode {
	return &CommentNode{
		node: newNode(nComment, indent, t),
		text: t.lit,
	}
}

func (n *CommentNode) Source(tw *templateWriter) error {
	if n.text != "" {
		return tw.WriteStringLiteral("<!--" + html.EscapeString(n.text) + "-->")
	}
	if err := tw.WriteStringLiteral("<!--"); err != nil {
		return err
	}
	for _, c := range n.children {
		if err := c.Source(tw); err != nil {
			return err
		}
	}
	return tw.WriteStringLiteral("-->")

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

func formattedText(text string) (string, bool) {
	if matches := reFmtText.FindStringSubmatch(text); matches != nil {
		return `hamlet.FormatString("` + matches[1] + `", ` + matches[2] + `)`, true
	}
	return text, false
}

func (n *TextNode) Source(tw *templateWriter) error {
	if n.isDynamic {
		vName := tw.GetVarName()
		if err := tw.WriteIndent(`var ` + vName + " string\n"); err != nil {
			return err
		}
		text, _ := formattedText(n.text) // does the code match a format string?

		if tw.isUnescaped {
			if err := tw.WriteIndent(`if ` + vName + `, __err = hamlet.CaptureErrors(` + text + "); __err != nil { return }\n"); err != nil {
				return err
			}
		} else {
			if err := tw.WriteIndent(`if ` + vName + `, __err = hamlet.CaptureErrors(hamlet.EscapeString(` + text + ")); __err != nil { return }\n"); err != nil {
				return err
			}
		}
		if err := tw.WriteStringIndent(vName); err != nil {
			return err
		}
		return nil
	}

	if n.isPlain || tw.isUnescaped {
		s := strconv.Quote(n.text)
		return tw.WriteStringLiteral(s[1 : len(s)-1])
	}

	if tw.isUnescaped {
		return tw.WriteStringLiteral(n.text)
	}

	return tw.WriteStringLiteral(html.EscapeString(n.text))
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
	code string
}

func NewSilentScriptNode(t token, indent int) *SilentScriptNode {
	return &SilentScriptNode{
		node: newNode(nSilentScriptNode, indent, t),
		code: t.lit,
	}
}

func (n *SilentScriptNode) Source(tw *templateWriter) error {
	code := strings.TrimSpace(n.code)

	if err := tw.WriteIndent(code + "\n"); err != nil {
		return err
	}

	if len(n.children) == 0 {
		return nil
	}

	itw := tw.Indent()
	for _, c := range n.children {
		if err := c.Source(itw); err != nil {
			return err
		}
	}
	if err := itw.Close(); err != nil {
		return err
	}

	return nil
}

func (n *SilentScriptNode) parse(p *parser) error {
	switch p.peek().Type() {
	case tNewLine:
		p.next()
		return nil
	default:
		return n.handleNode(p, n.indent+1)
	}
}

type ScriptNode struct {
	node
	code string
}

func NewScriptNode(t token) *ScriptNode {
	return &ScriptNode{
		node: newNode(nScriptNode, 0, t),
		code: t.lit,
	}
}

func (n *ScriptNode) Source(tw *templateWriter) error {

	vName := tw.GetVarName()
	if err := tw.WriteIndent(`var ` + vName + " string\n"); err != nil {
		return err
	}
	code, _ := formattedText(n.code)
	if tw.isUnescaped {
		if err := tw.WriteIndent(`if ` + vName + `, __err = hamlet.CaptureErrors(` + code + "); __err != nil { return }\n"); err != nil {
			return err
		}
	} else {
		if err := tw.WriteIndent(`if ` + vName + `, __err = hamlet.CaptureErrors(hamlet.EscapeString(` + code + ")); __err != nil { return }\n"); err != nil {
			return err
		}
	}
	if err := tw.WriteStringIndent(vName); err != nil {
		return err
	}
	return nil
}

type RenderCommandNode struct {
	node
	command string
}

func NewRenderCommandNode(t token, indent int) *RenderCommandNode {
	return &RenderCommandNode{
		node:    newNode(nRenderCommand, indent, t),
		command: t.lit,
	}
}

func (n *RenderCommandNode) Source(tw *templateWriter) error {
	vName := tw.GetVarName()

	fnLine := vName + " := hamlet.TemplateFunc(func(ctx context.Context, __w io.Writer) (__err error) {\n"

	if err := tw.WriteIndent(fnLine); err != nil {
		return err
	}

	itw := tw.Indent()

	lines := []string{
		"__buf, __isBuf := __w.(*bytes.Buffer)\n",
		"if !__isBuf {\n",
		"	__buf = hamlet.GetBuffer()\n",
		"	defer hamlet.ReleaseBuffer(__buf)\n",
		"}\n",
	}
	for _, line := range lines {
		if err := itw.WriteIndent(line); err != nil {
			return err
		}
	}
	for _, c := range n.children {
		if err := c.Source(itw); err != nil {
			return err
		}
	}
	if err := itw.Close(); err != nil {
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
		if err := tw.WriteIndent(line); err != nil {
			return err
		}
	}

	callLn := "if __err = " + n.command + ".Render(hamlet.PushChildren(ctx, " + vName + "), __buf); __err != nil { return }\n"
	if err := tw.WriteIndent(callLn); err != nil {
		return err
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
	return tw.WriteIndent("if __err = __children.Render(ctx, __buf); __err != nil { return }\n")
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
	if err := tw.WriteStringLiteral("<script>"); err != nil {
		return err
	}
	for _, c := range n.children {
		if err := c.Source(tw); err != nil {
			return err
		}
	}
	return tw.WriteStringLiteral("</script>")
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
	if err := tw.WriteStringLiteral("<style>"); err != nil {
		return err
	}
	for _, c := range n.children {
		if err := c.Source(tw); err != nil {
			return err
		}
	}
	return tw.WriteStringLiteral("</style>")
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
}

func NewTextFilterNode(t token, indent int) *TextFilterNode {
	return &TextFilterNode{
		node: newNode(nFilter, indent, t),
	}
}

func (n *TextFilterNode) Source(tw *templateWriter) error {
	for _, c := range n.children {
		if err := c.Source(tw); err != nil {
			return err
		}
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
