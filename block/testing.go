package block

import (
	"fmt"
)

// Used in unit-tests
// Returns true if the same

func compareRaw(a, b *Block) (bool, error) {
	if len(a.raw) != len(b.raw) {
		return false, fmt.Errorf(
			"len(a.raw) = %d, len(b.raw) = %d",
			len(a.raw),
			len(b.raw),
		)
	}

	for i := range a.raw {
		if len(a.raw[i]) != len(b.raw[i]) {
			return false, fmt.Errorf(
				"len(a.raw[%d]) = %d, len(b.raw[%d]) = %d",
				i, len(a.raw[i]),
				i, len(b.raw[i]),
			)
		}

		for j := range a.raw[i] {
			if a.raw[i][j] != b.raw[i][j] {
				return false, fmt.Errorf(
					"a.raw[%d][%d] = %s, b.raw[%d][%d] = %s",
					i, j, a.raw[i][j],
					i, j, b.raw[i][j],
				)
			}
		}
	}

	return true, nil
}

func CompareTreesRaw(a, b *Block) (bool, error) {
	res, err := compareRaw(a, b)
	if err != nil {
		return res, err
	}

	if !res {
		return res, nil
	}

	if len(a.children) != len(b.children) {
		return false, fmt.Errorf(
			"len(a.children) = %d, len(b.children) = %d",
			len(a.children),
			len(b.children),
		)
	}

	for i := range a.children {
		res, err := CompareTreesRaw(a.children[i], b.children[i])
		if err != nil {
			return false, err
		}

		if !res {
			return res, nil
		}
	}

	return true, nil
}
