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

var ErrUnexpectedToken = errors.New("unexpected token")
var ErrIllegalSymbolUsed = errors.New("reserved symbols used")
var ErrBlockNotEnclosed = errors.New("block is not enclosed")
var ErrFileExtension = errors.New("file extension must be .hoo")
var ErrSymbolsAfterBracket = errors.New("encountered symbols after \"{\"")
var ErrNamelessBlock = errors.New("block without name")

type Hoodie struct {
	scanner     *bufio.Scanner
	r           io.Reader
	srcPath     string
	outputPath  string
	raw         [][]string
	currentLine int
	head        *block.Block
}

func New(r io.Reader, outputPath, srcPath string) *Hoodie {

	return &Hoodie{
		scanner:    bufio.NewScanner(r),
		r:          r,
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
//	Create additional error types for various occasions.
func (h *Hoodie) SrcPath() string {
	return h.srcPath
}

// Performs initial read of the file into memory
// Check for balanced braces, removes whitespace and comments
func (h *Hoodie) Parse() error {
	var left, right int
	location := h.head // pointer to block being parsed
	for h.scan() {

		raw := h.scanner.Text()
		if strings.Index(raw, "//") != -1 {
			raw = raw[:strings.Index(raw, "//")]
		}

		line := lo.Compact(strings.FieldsFunc(raw, func(r rune) bool {
			return strings.ContainsRune(SPACE_TAB, r)
		}))

		if len(line) == 0 {
			continue
		}

		// TODO: head dosen't have `{`, can this cause issues?
		// 		head doesn't have it's own line in the file so maybe no
		if lo.Contains(line, "{") {
			left++
			b := block.New(h.srcPath, h.currentLine)
			b.WriteRaw(line)
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

		if lo.Contains(line, "}") {
			right++
			location = location.Parent()
			continue
		}

		location.WriteRaw(line)
	}

	if left != right {
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
TODO: use patterm matching from `mo` package?
	patterns:
		a "trait" "trait_name" "{"
			traits are literally never gonna be declared inside block why bother?
		b "word" "{"
		c "word" "trait_name" ... "{"
		d "word" "word"
		e "}"

		trat {} block as value and block name as a key?

Is order of declaring blocks and pairs significant?
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
