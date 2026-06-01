package main

type quadRune rune

func (r quadRune) toInt() int {
	switch r {
	case '▘':
		return 0b0001 // top-left
	case '▝':
		return 0b0010 // top-right
	case '▀':
		return 0b0011 // top
	case '▖':
		return 0b0100 // bottom-left
	case '▌':
		return 0b0101 // left
	case '▞':
		return 0b0110 // diagonal /
	case '▛':
		return 0b0111 // except bottom-right
	case '▗':
		return 0b1000 // bottom-right
	case '▚':
		return 0b1001 // diagonal \
	case '▐':
		return 0b1010 // right
	case '▜':
		return 0b1011 // except bottom-left
	case '▄':
		return 0b1100 // bottom
	case '▙':
		return 0b1101 // except top-right
	case '▟':
		return 0b1110 // except top-left
	case '█':
		return 0b1111 // all
	default:
		return 0
	}
}

func (r quadRune) update(offX, offY int, active, changed bool) quadRune {
	if changed {
		i := r.toInt()
		if active {
			i |= 1 << (offY*2 + offX)
		} else {
			i &^= 1 << (offY*2 + offX)
		}
		return quadCell[i]
	}
	return r
}

var quadCell = [16]quadRune{
	' ', // 0000
	'▘', // 0001 top-left
	'▝', // 0010 top-right
	'▀', // 0011 top
	'▖', // 0100 bottom-left
	'▌', // 0101 left
	'▞', // 0110 diagonal /
	'▛', // 0111 except bottom-right
	'▗', // 1000 bottom-right
	'▚', // 1001 diagonal \
	'▐', // 1010 right
	'▜', // 1011 except bottom-left
	'▄', // 1100 bottom
	'▙', // 1101 except top-right
	'▟', // 1110 except top-left
	'█', // 1111
}
