package html

import (
	"golang/net/html" // "golang.org/x/net/html"
	"golang/net/html/atom" // "golang.org/x/net/html/atom"
	"io"
	"strings"
)

const whitespace = " \t\r\n\f"

type Node struct {
	Parent *Node
	Children  []*Node
	Type      html.NodeType
	DataAtom  atom.Atom
	TagName      string
	Text        string
	Attr      []html.Attribute
	Close     bool
	ID        string
	Class     []string
	Nth       uint
}

type Parser struct {
	tokenizer *html.Tokenizer
	document  *Node
	token     html.Token
}


func ParseHtml(r io.Reader) (*Node, error) {

	p := &Parser{
		tokenizer: html.NewTokenizer(r),
		document: &Node{
			Type: html.DocumentNode,
		},
	}

	err := p.parse()

	if err != nil {

		return nil, err
	}

	return p.document, nil
}

func (p *Parser) parse() (err error){

	sequence := make([]int, 0, 10)

	for err != io.EOF {

		tokenType := p.tokenizer.Next()

		p.token = p.tokenizer.Token()

		child := new(Node)

		switch tokenType {

		case html.ErrorToken:

			err = p.tokenizer.Err()

			if err != nil && err != io.EOF {

				return err
			}

		case html.TextToken:

			p.token.Data = strings.TrimLeft(p.token.Data, whitespace)

			if len(p.token.Data) == 0 {

				continue
			}

			child = &Node{
				Type: html.TextNode,
				DataAtom: p.token.DataAtom,
				TagName: "text",
				Text: p.token.Data,
			}

		case html.StartTagToken:

			child = &Node{
				Type: html.ElementNode,
				DataAtom: p.token.DataAtom,
				TagName: p.token.Data,
				Attr: p.token.Attr,
			}

			tokenType = SelfClosing(child)

		case html.EndTagToken:

			child = &Node{
				TagName: p.token.Data,
			}


		case html.SelfClosingTagToken:

			child = &Node{
				Type: html.ElementNode,
				DataAtom: p.token.DataAtom,
				TagName: p.token.Data,
				Attr: p.token.Attr,
				Close: true,
			}

		case html.CommentToken:

			child = &Node{
				Type: html.CommentNode,
				DataAtom: p.token.DataAtom,
				TagName: "comment",
				Text: p.token.Data,
			}

		case html.DoctypeToken:

			child = &Node{
				Type: html.DoctypeNode,
				DataAtom: p.token.DataAtom,
				TagName: "doctype",
				Text: p.token.Data,
			}

		}

		sequence = p.document.AddChild(child, tokenType, sequence)

	}

	return nil
}

func (n *Node) AddChild(child *Node, tokenType html.TokenType, sequence []int) []int {

	var (
		node *Node = n
		tagName string = child.TagName
		)

	for _, index := range sequence {

		node = node.Children[index]
	}

	switch tokenType {

	case html.DoctypeToken, html.CommentToken, html.TextToken:

		child.Parent = node

		node.Children = append(node.Children, child)

	case html.StartTagToken, html.SelfClosingTagToken:

		for _, attr := range child.Attr {

			if attr.Key == "id" {

				child.ID = attr.Val
			}

			if attr.Key == "class" {

				attr.Val = strings.Replace(attr.Val, "  ", " ", -1)

				child.Class = strings.Split(attr.Val, " ")
			}

		}

		if tokenType == html.StartTagToken {

			sequence = append(sequence, len(node.Children))

		}

		child.Parent = node

		node.Children = append(node.Children, child)

	case html.EndTagToken:

		if node.TagName == child.TagName {

			node.Close = true

			sequence = sequence[:len(sequence) - 1]

			return sequence
		}

	case html.ErrorToken:

		return sequence
	}

	for _, c := range child.Parent.Children {

		if c.TagName == tagName {

			child.Nth = child.Nth + 1
		}

	}

	return sequence
	
}


func SelfClosing(node *Node) html.TokenType {

	switch node.TagName {

	case "br", "hr", "img", "input", "param", "meta", "link", "area", "base", "col", "command", "embed", "keygen", "source", "track", "wbr":

		node.Close = true

		return html.SelfClosingTagToken
	}

	return html.StartTagToken

}

