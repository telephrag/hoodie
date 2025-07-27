package block

import (
	"fmt"
	"strings"
)

type BlockError struct {
	path [][]string
	err  error
}

func (e BlockError) Error() string {

	errStr := fmt.Sprintf("%s: %s\n", e.path[0][0], e.err)

	depth := 1
	for _, hop := range e.path[1:] {
		name := "(" + strings.Join(hop, " ") + ")"
		pad := strings.Repeat("\u0020\u0020\u0020L", depth)
		errStr = errStr + pad + name + "\n"
	}
	return errStr
}

func (b *Block) Err(err error) error {

	path := make([][]string, 2)
	path[0] = []string{b.srcPath}
	path[1] = b.name

	parent := b.parent
	for !parent.isHead {
		path = append(path, parent.name)
		parent = parent.parent
	}

	return BlockError{path, err}
}

// It's massive skill issue that I can't parse it in a single pass...

// I want to provide correct line # in error even when block
// has children inbetween pairs.
// 	> I have to pass more data into `Err()`
//  	to deduce where an error has occured.
// 	> Or to use block's state to infer it somehow so, I would have to enforce
//    	state just to provide information about error
//  > Or just disallow pairs after child blocks

// Safe to assume the children list is complete, so:
// 1. Jump to all the children headers
// 2. Traverse until {} are balanced
// 	> this assumes that in hoodie.Parse() we've already
// 		errored on unbalanced braces if there were any
// 3. Remember the amount of lines traversed
// 4. Find a line where an error has occured
//
// We might be able to simplify it in some cases cause
// children won't always be sitting inbetween blocks

// Provide end line in addition to start line?

// We still have commentaries...., so better not bother with line #
