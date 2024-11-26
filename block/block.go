package block

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/samber/lo"
)

var ErrBadHeader = errors.New("bad block header")
var ErrBadTraitHeader = errors.New("bad trait header")
var ErrTraitNests = errors.New("trait nests other blocks")
var ErrTraitNested = errors.New("trait is nested inside another block")
var ErrNotPair = errors.New("not a pair")
var ErrTraitExists = errors.New("trait already exists (name is not unique)")
var ErrBadConditional = errors.New("bad conditional")

var traits map[string]*Block = make(map[string]*Block)

type Block struct {
	name      []string // we'll store name and all traitsA
	srcPath   string
	startLine int
	line      int
	raw       [][]string
	contents  map[string]string
	parsed    bool
	head      bool
	parent    *Block
	children  []*Block
}

func NewHead(srcPath string) *Block {
	b := New(srcPath, 1)
	b.head = true
	b.name = []string{"jïr"}
	return b
}

func New(srcPath string, startLine int) *Block {
	b := &Block{}
	b.startLine = startLine
	b.line = startLine
	b.raw = make([][]string, 0)
	b.contents = map[string]string{}
	b.children = make([]*Block, 0)
	return b
}

func (b *Block) IsHead() bool { return b.head }

func (b *Block) Header() string {
	if len(b.raw) != 0 {
		return fmt.Sprint(b.raw[0])
	}
	return ""
}

func (b *Block) Name() []string {
	return b.name
}

func (b *Block) Parent() *Block {
	return b.parent
}

func (b *Block) AttachChild(child *Block, line int) *Block {
	child.parent = b
	b.children = append(b.children, child)
	return child
}

func (b *Block) WriteRaw(raw []string) {
	b.raw = append(b.raw, raw)
}

func (b *Block) Add(other *Block) error {
	// TODO: other.Parse() will run checks that we don't need when parsing traits
	if err := other.Parse(); err != nil {
		return err
	}

	for k, v := range other.contents {
		b.contents[k] = v
	}

	return nil
}

// Called when `{` is encountered by `hoodie.Parse()`
func (b *Block) ParseHeader() (isTrait bool, err error) {

	name := b.raw[0]
	if len(name) < 2 { // expecting at least name and `{`
		return false, ErrBadHeader
	}

	if isTrait := (name[0] == "trait"); isTrait {

		if len(name) != 3 { // expecting `trait`, name, `{`
			return false, ErrBadTraitHeader
		}
		name = name[1:] // excluding `trait` keyword
		// TODO: Should this logic be here?
		if _, ok := traits[name[0]]; ok {
			return false, ErrTraitExists
		}
		traits[name[0]] = b // adding trait to the table
		isTrait = true
	}

	b.name = name[:len(name)-1] // excluding `{`

	b.line++

	return isTrait, nil
}

func parsePair(line []string) (string, string, error) {
	if len(line) != 2 {
		return "", "", ErrNotPair
	}
	return line[0], line[1], nil
}

func (b *Block) Parse() error {
	if b.parsed || b.head { // lazy eval for traits, not parsing head
		return nil
	}
	defer func() { b.parsed = true }()

	contents := make(map[string]string)
	for _, line := range b.raw[1:] {
		left, right, err := parsePair(line)
		if err != nil {
			return b.Err(err)
		}
		contents[left] = right
		b.line++ // TODO: line # won't be correct if `b` has children (???)
	}

	if len(b.name) > 1 {
		for _, traitName := range b.name[1:] {
			if err := b.Add(traits[traitName]); err != nil {
				return err
			}
		}
		for k, v := range contents {
			b.contents[k] = v
		}
	} else {
		b.contents = contents
	}

	b.line = b.startLine

	return nil
}

func ParseTree(root *Block) error {
	if err := root.Parse(); err != nil {
		return err
	}

	for _, c := range root.children {
		if err := ParseTree(c); err != nil {
			return err
		}
	}

	return nil
}

func (b *Block) RemoveTraitsFromChildren() {
	lo.Filter(b.children, func(x *Block, i int) bool {
		_, ok := traits[b.children[i].name[0]]
		return ok
	})
}

func (b *Block) Compile(w io.Writer, depth int) {

	fmt.Fprintf(w, "\n%s\"%s\"\n", // header
		strings.Repeat("\u0009", depth),
		b.name[0],
	)
	fmt.Fprint(w, strings.Repeat("\u0009", depth), "{\n")
	b.line++ // we count lines as they are in .hoo file

	for k, v := range b.contents { // pairs
		var condTag string
		if condIndex := strings.LastIndex(k, "$"); condIndex != -1 {
			condTag = fmt.Sprintf("[%s]", k[condIndex:])
			k = k[:condIndex]
		}
		fmt.Fprintf(w, "%s\"%s\" \"%s\" %s\n",
			strings.Repeat("\u0009", depth+1),
			k, v, condTag,
		)
		b.line++
	}

	b.line = b.startLine
}

func CompileTree(root *Block, w io.Writer, depth int) error {
	// not writing head as it's here only for technical reasons
	for _, c := range root.children {
		c.Compile(w, depth)

		if err := CompileTree(c, w, depth+1); err != nil {
			return err
		}

		fmt.Fprintf(w, "%s}\n", strings.Repeat("\u0009", depth))
	}

	return nil
}

func PrintTree(root *Block, depth int) {
	fmt.Printf("%s%s:\n", strings.Repeat("    ", depth), root.name[0])
	for k, v := range root.contents {
		fmt.Printf("%s%s : %s\n", strings.Repeat("    ", depth+1), k, v)
	}
	fmt.Println()
	for _, c := range root.children {
		PrintTree(c, depth+1)
	}
}

func ValidateTrates() error {
	for _, trait := range traits {
		if len(trait.children) != 0 {
			return trait.Err(ErrTraitNests)
		}
		if !trait.Parent().head {
			return trait.Err(ErrTraitNested)
		}
	}

	return nil
}
