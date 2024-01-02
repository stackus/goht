package compiler

import (
	"fmt"
	"os"
)

type parser struct {
	lexer    *lexer
	template *Template
	n        parsingNode
	nodes    stack[parsingNode]
	token    token
	tokens   stack[token]
}

func ParseFile(fileName string) (*Template, error) {
	contents, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	return parseBytes(contents)
}

func ParseString(contents string) (*Template, error) {
	return parseBytes([]byte(contents))
}

func parseBytes(contents []byte) (*Template, error) {
	p := newParser(contents)

	err := p.parse()

	return p.template, err
}

func newParser(contents []byte) *parser {
	rootNode := NewRootNode()

	nodes := stack[parsingNode]{rootNode}

	p := &parser{
		lexer: newLexer(contents),
		template: &Template{
			Root: rootNode,
		},
		n:     rootNode,
		nodes: nodes,
	}

	return p
}

func (p *parser) nextToken() {
	token := p.lexer.nextToken()
	p.tokens.push(token)
}

func (p *parser) parse() error {
	p.nextToken()
	for {
		if err := p.n.parse(p); err != nil {
			return err
		}
		if p.token.Type() == tEOF {
			break
		}
	}
	return nil
}

func (p *parser) peek() token {
	return p.tokens.peek()
}

func (p *parser) next() token {
	p.token = p.tokens.pop()
	p.nextToken()
	return p.token
}

func (p *parser) backToType(typ nodeType) error {
	for p.nodes.peek().Type() != typ {
		if p.nodes.peek().Type() == nRoot {
			return fmt.Errorf("unexpected: node has no parent %d", p.nodes.peek().Type())
		}
		p.nodes.pop()
	}
	p.n = p.nodes.peek()
	return nil
}

func (p *parser) backToIndent(indent int) error {
	for {
		if p.nodes.peek().Type() == nRoot {
			return fmt.Errorf("unexpected: node has no parent %d", p.nodes.peek().Type())
		}
		if p.nodes.peek().Indent() <= indent {
			break
		}
		p.nodes.pop()
	}
	p.n = p.nodes.peek()
	return nil
}

func (p *parser) backToParent() error {
	if p.nodes.peek().Type() == nRoot {
		return fmt.Errorf("unexpected: node has no parent %d", p.nodes.peek().Type())
	}
	p.nodes.pop()
	p.n = p.nodes.peek()
	return nil
}

func (p *parser) addNode(n parsingNode) {
	p.addChild(n)
	p.nodes.push(n)
	p.n = n
}

func (p *parser) addChild(n nodeBase) {
	p.n.AddChild(n)
}
