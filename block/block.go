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

// var ErrBadConditional = errors.New("bad conditional")

var traits map[string]*Block = make(map[string]*Block)

type Block struct {
	name      []string // name and all traits
	srcPath   string
	rawTokens [][]string
	contents  map[string]string
	parsed    bool
	isHead    bool
	parent    *Block
	children  []*Block
}

func NewHead(srcPath string) *Block {
	b := New(srcPath)
	b.name = []string{"jïr"}
	b.isHead = true
	return b
}

func New(srcPath string) *Block {
	b := &Block{}
	b.srcPath = srcPath
	b.rawTokens = make([][]string, 0)
	b.contents = map[string]string{}
	b.children = make([]*Block, 0)
	return b
}

func (b *Block) IsHead() bool { return b.isHead }

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

// TODO: Perhaps read line-count of file and pre-allocate
func (b *Block) WriteRaw(raw []string) {
	b.rawTokens = append(b.rawTokens, raw)
}

func (b *Block) Add(other *Block) error {
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

	tokens := b.rawTokens[0]
	if len(tokens) < 2 { // expecting at least name and `{`
		return false, ErrBadHeader
	}

	if isTrait := (tokens[0] == "trait"); isTrait {

		if len(tokens) != 3 { // expecting `trait`, name, `{`
			return false, ErrBadTraitHeader
		}

		tokens = tokens[1:] // excluding `trait` keyword

		if _, ok := traits[tokens[0]]; ok {
			return false, ErrTraitExists
		}
		traits[tokens[0]] = b // adding trait to the table
		isTrait = true
	}

	if tokens[len(tokens)-1] != "{" {
		return isTrait, ErrBadHeader
	}

	b.name = tokens[:len(tokens)-1] // excluding `{`

	return isTrait, nil
}

func parsePair(line []string) (string, string, error) {
	if len(line) != 2 {
		return "", "", ErrNotPair
	}
	return line[0], line[1], nil
}

func (b *Block) Parse() error {
	if b.parsed || b.isHead { // not re-parsing the same trait
		return nil
	}
	defer func() { b.parsed = true }()

	contents := make(map[string]string)
	for _, line := range b.rawTokens[1:] {
		// TODO: provide correct line # on error
		left, right, err := parsePair(line)
		if err != nil {
			return b.Err(err)
		}
		contents[left] = right
		// TODO: line # won't be correct if `b` has children between pairs
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

	fmt.Fprintf(w, "%s\"%s\"\n", // header
		strings.Repeat("\u0009", depth),
		b.name[0],
	)
	fmt.Fprint(w, strings.Repeat("\u0009", depth), "{\n")

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
	}
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
		if !trait.Parent().isHead {
			return trait.Err(ErrTraitNested)
		}
	}

	return nil
}
