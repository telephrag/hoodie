package hoodie

import "fmt"

// This is exactly the same type as in `block`.
// It is small enough to warrant mere copypaste
// instead of it's own dependency.

type HoodieErr struct {
	srcPath string
	line    int
	err     error
}

func (e HoodieErr) Error() string {
	return fmt.Sprintf("%s, line %d: %s\n", e.srcPath, e.line, e.err)
}

func (h *Hoodie) Err(err error) HoodieErr {
	return HoodieErr{h.srcPath, h.currentLine, err}
}
