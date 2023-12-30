package transpiler

import (
	"bytes"
	"fmt"
	"html"
	"slices"
	"strconv"
	"strings"
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
	nText
	nUnescape
	nExecuteCode
	nOutputCode
	nRenderCommand
	nChildrenCommand
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
	case nText:
		return "Text"
	case nUnescape:
		return "Unescape"
	case nExecuteCode:
		return "ExecuteCode"
	case nOutputCode:
		return "OutputCode"
	case nRenderCommand:
		return "RenderCommand"
	case nChildrenCommand:
		return "ChildrenCommand"
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
	Origin() tokenType
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
	origin   tokenType
	children []nodeBase
}

func newNode(typ nodeType, indent int, origin tokenType) node {
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

func (n *node) Origin() tokenType {
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

func (n *node) handleNode(p *parser, indent int) error {
	switch p.peek().Type() {
	case tNewLine:
		p.next()
	case tIndent:
		nextIndent := len(p.next().Lit())
		if nextIndent <= n.indent {
			return p.backToIndent(nextIndent - 1)
		}
		return n.handleNode(p, nextIndent)
	case tDoctype:
		p.addChild(NewDoctypeNode(p.next()))
	case tTag, tId, tClass:
		p.addNode(NewElementNode(p.next(), indent))
	case tUnescaped:
		p.addNode(NewUnescapeNode(p.next(), indent))
	case tStaticText, tDynamicText:
		p.addChild(NewTextNode(p.next()))
	case tExecuteCode:
		p.addNode(NewExecuteCodeNode(p.next(), indent))
	case tOutputCode:
		p.addChild(NewOutputCodeNode(p.next()))
	case tRenderCommand:
		p.addNode(NewRenderCommandNode(p.next(), indent))
	case tChildrenCommand:
		p.addChild(NewChildrenCommandNode(p.next()))
	case tHmltEnd:
		return p.backToType(nHmlt)
	case tEOF:
		return fmt.Errorf("%s: template is incomplete: %s", n.typ, p.peek())
	default:
		return fmt.Errorf("%s: unexpected: %s", n.typ, p.peek())
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
		node: newNode(nRoot, 0, tRoot),
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
		n.pkg = p.next().Lit()
	case tImport:
		t := p.next()
		if !slices.Contains(n.imports, t.Lit()) {
			n.imports = append(n.imports, t.Lit())
		}
	case tGoCode, tNewLine:
		p.addNode(NewCodeNode(p.next()))
	case tHmltStart:
		p.addNode(NewHmltNode(p.next()))
	case tEOF:
		p.next()
		return nil
	default:
		return fmt.Errorf("%s: unexpected: %s", n.Type(), p.peek())
	}

	return nil
}

type CodeNode struct {
	node
	text *strings.Builder
}

func NewCodeNode(t token) *CodeNode {
	builder := &strings.Builder{}
	builder.WriteString(t.Lit())
	return &CodeNode{
		node: newNode(nGoCode, 0, t.Type()),
		text: builder,
	}
}

func (n *CodeNode) Source(tw *templateWriter) error {
	return tw.Write(n.text.String())
}

func (n *CodeNode) parse(p *parser) error {
	switch p.peek().Type() {
	case tGoCode, tNewLine:
		t := p.next()
		_, err := n.text.WriteString(t.Lit())
		return err
	case tImport, tHmltStart, tEOF:
		return p.backToType(nRoot)
	default:
		return fmt.Errorf("%s: unexpected: %s", n.Type(), p.peek())
	}
}

type HmltNode struct {
	node
	decl string
}

func NewHmltNode(t token) *HmltNode {
	return &HmltNode{
		node: newNode(nHmlt, -1, t.Type()),
		decl: t.Lit(),
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
			_, __err = __buf.WriteTo(__w)
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
		node:    newNode(nDoctype, 0, t.Type()),
		doctype: t.Lit(),
	}
}

func (n *DoctypeNode) Source(tw *templateWriter) error {
	doctype := ""
	switch n.doctype {
	case "":
		doctype = `<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML 1.0 Transitional//EN\" \"http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd\">`
	case "Strict", "strict":
		doctype = `<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML 1.0 Strict//EN\" \"http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd\">`
	case "Frameset", "frameset":
		doctype = `<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML 1.0 Frameset//EN\" \"http://www.w3.org/TR/xhtml1/DTD/xhtml1-frameset.dtd\">`
	case "5":
		doctype = `<!DOCTYPE html>`
	case "1.1":
		doctype = `<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML 1.1//EN\" \"http://www.w3.org/TR/xhtml11/DTD/xhtml11.dtd\">`
	case "Basic", "basic":
		doctype = `<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML Basic 1.1//EN\" \"http://www.w3.org/TR/xhtml-basic/xhtml-basic11.dtd\">`
	case "Mobile", "mobile":
		doctype = `<!DOCTYPE html PUBLIC \"-//WAPFORUM//DTD XHTML Mobile 1.2//EN\" \"http://www.openmobilealliance.org/tech/DTD/xhtml-mobile12.dtd\">`
	case "RDFa", "rdfa":
		doctype = `<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML+RDFa 1.0//EN\" \"http://www.w3.org/MarkUp/DTD/xhtml-rdfa-1.dtd\">`
	default:
		return fmt.Errorf("%s: unknown doctype: %s", n.Type(), n.doctype)
	}
	return tw.WriteStringLiteral(doctype)
}

type attribute struct {
	name      string
	isBoolean bool
	isDynamic bool
	value     string
}

type ElementNode struct {
	node
	tag              string
	id               string
	classes          []string
	attributes       *OrderedMap[attribute]
	attributesCmd    string
	disallowChildren bool
	isSelfClosing    bool
	isComplete       bool // TODO might be a better way of tracking if we're done parsing the element
}

func NewElementNode(t token, indent int) *ElementNode {
	n := &ElementNode{
		node:       newNode(nElement, indent, t.Type()),
		tag:        "div",
		attributes: NewOrderedMap[attribute](),
	}

	switch t.Type() {
	case tTag:
		n.tag = t.Lit()
	case tId:
		n.id = t.Lit()
	case tClass:
		n.classes = append(n.classes, fmt.Sprintf("%q", t.Lit()))
	}

	return n
}

func (n *ElementNode) Source(tw *templateWriter) error {
	if err := tw.WriteStringLiteral("<" + n.tag); err != nil {
		return err
	}
	// write attributes
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
			if err := tw.Indent().WriteStringLiteral(" " + attr.name); err != nil {
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
			if err := tw.WriteStringIndent(`hamlet.EscapeString(` + attr.value + `)+"\""`); err != nil {
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
	if err := tw.WriteStringLiteral(">"); err != nil {
		return err
	}
	if n.isSelfClosing {
		return nil
	}

	for _, c := range n.children {
		if err := c.Source(tw); err != nil {
			return err
		}
	}
	if err := tw.WriteStringLiteral("</" + n.tag + ">"); err != nil {
		return err
	}

	return nil
}

func (n *ElementNode) renderClass(tw *templateWriter) error {
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
	if n.isComplete {
		switch p.peek().Type() {
		case tIndent:
			nextIndent := len(p.peek().Lit())
			if nextIndent <= n.indent {
				return p.backToIndent(nextIndent - 1)
			}
			if nextIndent > n.indent && n.disallowChildren {
				return fmt.Errorf("%s: illegal nesting: %s", n.Type(), p.token)
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
	case tId:
		n.id = html.EscapeString(p.next().Lit())
	case tClass:
		n.classes = append(n.classes, strconv.Quote(html.EscapeString(p.next().Lit())))
	case tAttrName:
		return n.parseAttributes(p)
	case tAttributesCommand:
		n.attributesCmd = p.next().Lit()
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

		name := p.next().Lit()
		switch p.peek().Type() {
		case tAttrName:
			n.attributes.Set(name, attribute{
				name: name,
			})
			continue
		case tAttrOperator:
			op := p.next().Lit()
			switch op {
			case "?:":
				isBoolean = true
				if p.peek().Type() != tAttrDynamicValue {
					return fmt.Errorf("%s: expected dynamic value: %s", n.Type(), p.peek())
				}
			}
		}
		if p.peek().Type() != tAttrDynamicValue && p.peek().Type() != tAttrStaticValue {
			return fmt.Errorf("%s: expected attribute value: %s", n.Type(), p.peek())
		}
		if p.peek().Type() == tAttrDynamicValue {
			isDynamic = true
			value = p.next().Lit()
		} else {
			value, _ = strconv.Unquote(p.next().Lit())
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

	n.attributes.Range(func(_ string, attr attribute) (bool, error) {
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
	if n.attributesCmd != "" {
		attrs = append(attrs, "@attrs={"+n.attributesCmd+"...}")
	}

	buf.WriteString(lead + n.Type().String() + " " + n.tag + "(" + strings.Join(attrs, ",") + ")\n")
	for _, c := range n.children {
		c.Tree(buf, indent+1)
	}
	return buf.String()
}

type TextNode struct {
	node
	text      string
	isDynamic bool
}

func NewTextNode(t token) *TextNode {
	return &TextNode{
		node:      newNode(nText, 0, t.Type()),
		text:      t.Lit(),
		isDynamic: t.Type() == tDynamicText,
	}
}

func (n *TextNode) Source(tw *templateWriter) error {
	if n.isDynamic {
		vName := tw.GetVarName()
		if err := tw.WriteIndent(`var ` + vName + " string\n"); err != nil {
			return err
		}
		if tw.isUnescaped {
			if err := tw.WriteIndent(`if ` + vName + `, __err = hamlet.CaptureErrors(` + n.text + "); __err != nil { return }\n"); err != nil {
				return err
			}
		} else {
			if err := tw.WriteIndent(`if ` + vName + `, __err = hamlet.CaptureErrors(hamlet.EscapeString(` + n.text + ")); __err != nil { return }\n"); err != nil {
				return err
			}
		}
		if err := tw.WriteStringIndent(vName); err != nil {
			return err
		}
		return nil
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
		node: newNode(nUnescape, indent, t.Type()),
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
	switch p.peek().Type() {
	case tNewLine:
		return p.backToParent()
	default:
		return n.handleNode(p, n.indent)
	}
}

type ExecuteCodeNode struct {
	node
	code string
}

func NewExecuteCodeNode(t token, indent int) *ExecuteCodeNode {
	return &ExecuteCodeNode{
		node: newNode(nExecuteCode, indent, t.Type()),
		code: t.Lit(),
	}
}

func (n *ExecuteCodeNode) Source(tw *templateWriter) error {
	if err := tw.WriteIndent(n.code); err != nil {
		return err
	}

	if len(n.children) == 0 {
		return nil
	}

	if !strings.HasSuffix(strings.TrimSpace(n.code), "{") {
		if err := tw.Write(" {\n"); err != nil {
			return err
		}
	} else {
		if err := tw.Write("\n"); err != nil {
			return err
		}
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

	if err := tw.WriteIndent("}\n"); err != nil {
		return err
	}

	return nil
}

func (n *ExecuteCodeNode) parse(p *parser) error {
	return n.handleNode(p, n.indent+1)
}

type OutputCodeNode struct {
	node
	code string
}

func NewOutputCodeNode(t token) *OutputCodeNode {
	return &OutputCodeNode{
		node: newNode(nOutputCode, 0, t.Type()),
		code: t.Lit(),
	}
}

func (n *OutputCodeNode) Source(tw *templateWriter) error {
	vName := tw.GetVarName()
	if err := tw.WriteIndent(`var ` + vName + " string\n"); err != nil {
		return err
	}
	if tw.isUnescaped {
		if err := tw.WriteIndent(`if ` + vName + `, __err = hamlet.CaptureErrors(` + n.code + "); __err != nil { return }\n"); err != nil {
			return err
		}
	} else {
		if err := tw.WriteIndent(`if ` + vName + `, __err = hamlet.CaptureErrors(hamlet.EscapeString(` + n.code + ")); __err != nil { return }\n"); err != nil {
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
		node:    newNode(nRenderCommand, indent, t.Type()),
		command: t.Lit(),
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

	callLn := "if __err = " + n.command + ".Render(hamlet.PushChildren(ctx, " + vName + "), __w); __err != nil { return }\n"
	if err := tw.WriteIndent(callLn); err != nil {
		return err
	}

	return nil
}

func (n *RenderCommandNode) parse(p *parser) error {
	return n.handleNode(p, n.indent+1)
}

type ChildrenCommandNode struct {
	node
}

func NewChildrenCommandNode(t token) *ChildrenCommandNode {
	return &ChildrenCommandNode{
		node: newNode(nChildrenCommand, 0, t.Type()),
	}
}

func (n *ChildrenCommandNode) Source(tw *templateWriter) error {
	return tw.WriteIndent("if __err = __children.Render(ctx, __buf); __err != nil { return }\n")
}
