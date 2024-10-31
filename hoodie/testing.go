package hoodie

import "main/block"

func CompareTreesRaw(a, b *Hoodie) (bool, error) {
	return block.CompareTreesRaw(a.head, b.head)
}
