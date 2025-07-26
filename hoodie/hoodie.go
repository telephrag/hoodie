package hoodie

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"main/block"
	"os"
	"strings"

	"github.com/samber/lo"
)

const SPACE_TAB = "\u0020\u0009"

// TODO: most of these are unused, see comment bellow
var ErrUnexpectedToken = errors.New("unexpected token")
var ErrIllegalSymbolUsed = errors.New("reserved symbols used")
var ErrBlockNotEnclosed = errors.New("block is not enclosed")
var ErrSymbolsAfterBracket = errors.New("encountered symbols after \"{\"")

type Hoodie struct {
	scanner     *bufio.Scanner
	srcPath     string
	outputPath  string
	raw         [][]string
	currentLine int
	head        *block.Block
}

func New(r io.Reader, outputPath, srcPath string) *Hoodie {

	return &Hoodie{
		scanner:    bufio.NewScanner(r),
		srcPath:    srcPath,
		outputPath: outputPath,
		raw:        make([][]string, 0),
		head:       block.NewHead(srcPath),
	}
}

func (h *Hoodie) scan() bool {
	h.currentLine++
	return h.scanner.Scan()
}

// TODO: This is incredibly ugly. Rework error handling.
//
//	Create additional error types for various occasions. See above.
func (h *Hoodie) SrcPath() string {
	return h.srcPath
}

// Performs initial read of the file into memory;
// checks for balanced braces, removes whitespace and comments.
// Each file has a tree of blocks with imaginary head as a root.
func (h *Hoodie) Parse() error {
	var leftCurly, rightCurly int // need even count of these
	location := h.head            // parse from the root
	for h.scan() {
		raw := h.scanner.Text()

		// split will be to slow, slice with index
		if si := strings.Index(raw, "//"); si != -1 {
			raw = raw[:si]
		}

		// strings.Split will give us empty lines so, we use `lo`
		tokens := lo.Compact(strings.FieldsFunc(raw, func(r rune) bool {
			return strings.ContainsRune(SPACE_TAB, r)
		}))

		if len(tokens) == 0 {
			continue
		}

		// TODO: head dosen't have `{`, can this cause issues?
		// 		head doesn't have it's own line in the file so maybe no
		if lo.Contains(tokens, "{") {
			leftCurly++
			b := block.New(h.srcPath, h.currentLine)
			b.WriteRaw(tokens)
			// Parsing headers ahead of time for lazy trait evaluation
			isTrait, err := b.ParseHeader()
			if err != nil {
				return h.Err(err)
			}
			// trait can't have a parent except head
			if !isTrait || location.IsHead() {
				location = location.AttachChild(b, h.currentLine)
			}

			continue
		}

		if lo.Contains(tokens, "}") {
			rightCurly++
			location = location.Parent()
			continue
		}

		location.WriteRaw(tokens)
	}

	if leftCurly != rightCurly {
		return h.Err(ErrBlockNotEnclosed)
	}

	h.currentLine = 0
	return nil
}

func (h *Hoodie) ParseHead() error {
	return block.ParseTree(h.head)
}

func (h *Hoodie) WriteOutput() error {
	// files without this output path don't need to be written
	if h.outputPath == "_" {
		return nil
	}

	f, err := os.Create(h.outputPath)
	if err != nil {
		return fmt.Errorf(
			"failed to open %s for writing output: %w\n",
			h.outputPath, err,
		)
	}

	h.head.RemoveTraitsFromChildren()

	if err := block.CompileTree(h.head, f, 0); err != nil {
		return err
	}

	return nil
}

func (h *Hoodie) PrintTree() {
	block.PrintTree(h.head, 0)
}

/*
TODO: Is order of declaring blocks and pairs significant?
If yes, conditionals (e.g. if_match) can cause issues
Need more info if we're to allow nested blocks inside traits

{ we just need even amounf of left and right braces
	{}
	{
		{}{}
		{
			{}{}
		}
	}
}
*/
