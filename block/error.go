package block

import "fmt"

type BlockError struct {
	srcPath string
	line    int
	err     error
}

func (e BlockError) Error() string {
	return fmt.Sprintf("%s, line %d: %s\n", e.srcPath, e.line, e.err)
}

func (b *Block) Err(err error) error {
	return BlockError{b.srcPath, b.line, err}
}
