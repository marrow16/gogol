package patterns

import (
	"bufio"
	"errors"
	"github.com/marrow16/gogol/logic"
	"io"
	"strconv"
	"strings"
)

func PatternRleDecoder(r io.Reader) (result Pattern, err error) {
	scanner := bufio.NewScanner(r)
	var dataStarted, endSeen bool
	var data strings.Builder
	for scanner.Scan() && !endSeen {
		line := scanner.Text()
		switch {
		case dataStarted:
			if strings.HasSuffix(line, "!") {
				endSeen = true
			}
			data.WriteString(line)
		case strings.HasPrefix(line, "#N"):
			result.Name = strings.TrimSpace(line[2:])
		case strings.HasPrefix(line, "#C") || strings.HasPrefix(line, "#c"):
			result.Comments = append(result.Comments, strings.TrimSpace(line[2:]))
		case strings.HasPrefix(line, "#O"):
			result.Origination = strings.TrimSpace(line[2:])
		case strings.HasPrefix(line, "#P") || strings.HasPrefix(line, "#R"):
			result.Coordinates = strings.TrimSpace(line[2:])
		case strings.HasPrefix(line, "#r"):
			result.Rule, err = logic.NewRuleRle("", strings.TrimSpace(line[2:]))
			if err != nil {
				return
			}
		case strings.HasPrefix(line, "x") || strings.HasPrefix(line, "y"):
			parts := strings.Split(line, ",")
			if len(parts) < 2 {
				err = errors.New("invalid RLE format - bad dimension line")
				return
			}
			for _, p := range parts {
				dims := strings.Split(strings.TrimSpace(p), "=")
				if len(dims) == 2 {
					switch strings.TrimSpace(dims[0]) {
					case "x":
						var n int
						if n, err = strconv.Atoi(strings.TrimSpace(dims[1])); err == nil {
							result.Width = n
						} else {
							err = errors.New("invalid RLE format - bad dimension line")
						}
					case "y":
						var n int
						if n, err = strconv.Atoi(strings.TrimSpace(dims[1])); err == nil {
							result.Height = n
						} else {
							err = errors.New("invalid RLE format - bad dimension line")
						}
					case "rule":
						rle := strings.TrimSpace(strings.ToUpper(dims[1]))
						if cAt := strings.Index(rle, ":"); cAt != -1 {
							rle = rle[:cAt]
						}
						if !strings.ContainsRune(rle, 'S') && !strings.ContainsRune(rle, 'B') {
							// fix up lazy rules - e.g. "23/3" is "B3/S23"
							if bs := strings.Split(rle, "/"); len(bs) == 2 {
								rle = "B" + strings.TrimSpace(bs[1]) + "/S" + strings.TrimSpace(bs[0])
							}
						}
						result.Rule, err = logic.NewRuleRle("", rle)
						if err != nil {
							return
						}
					}
				}
			}
			if result.Width < 1 || result.Height < 1 {
				err = errors.New("invalid RLE format - bad dimension line")
				return
			}
			dataStarted = true
		}
	}
	if !dataStarted {
		err = errors.New("invalid RLE format - no data")
		return
	} else if !endSeen {
		err = errors.New("invalid RLE format - no end")
		return
	}
	result.Cells, err = ParseRLEData(result.Width, result.Height, data.String())
	return
}

func ParseRLEData(width, height int, data string) ([]bool, error) {
	result := make([]bool, width*height)
	row, col := 0, 0
	run := ""
	for _, ch := range []byte(data) {
		switch ch {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			run += string(ch)
			continue
		case ' ', '\t', '\r', '\n':
			continue
		case '!':
			if run != "" {
				return nil, errors.New("invalid RLE format - dangling run length")
			}
			return result, nil
		}
		n := 1
		if run != "" {
			var err error
			if n, err = strconv.Atoi(run); err != nil || n < 1 {
				return nil, errors.New("invalid RLE format - bad run length")
			}
			run = ""
		}
		switch ch {
		case 'o', 'b', 'x', 'y', 'z':
			if col+n > width {
				return nil, errors.New("invalid RLE format - row exceeds width")
			}
			if ch != 'b' {
				start := row*width + col
				for i := 0; i < n; i++ {
					result[start+i] = true
				}
			}
			col += n
		case '$':
			row += n
			col = 0
			if row >= height {
				return nil, errors.New("invalid RLE format - too many rows")
			}
		default:
			return nil, errors.New("invalid RLE format")
		}
	}
	if run != "" {
		return nil, errors.New("invalid RLE format - dangling run length")
	}
	return result, nil
}
