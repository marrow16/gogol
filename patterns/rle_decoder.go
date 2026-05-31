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
				line = strings.TrimSuffix(line, "!")
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
						if n, err := strconv.Atoi(strings.TrimSpace(dims[1])); err == nil {
							result.Width = n
						}
					case "y":
						if n, err := strconv.Atoi(strings.TrimSpace(dims[1])); err == nil {
							result.Height = n
						}
					case "rule":
						result.Rule, err = logic.NewRuleRle("", strings.TrimSpace(dims[1]))
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
	dataRows := strings.Split(data.String(), "$")
	if len(dataRows) > result.Height {
		err = errors.New("invalid RLE format - too many rows")
		return
	}
	result.Cells, err = parseRows(result.Width, result.Height, dataRows)
	return
}

func parseRows(width, height int, rows []string) (result []bool, err error) {
	mx := width * height
	result = make([]bool, mx)
	for r, row := range rows {
		rl := ""
		start := r * width
		for _, ch := range []byte(row) {
			switch ch {
			case 'o':
				// alive
				n := 1
				if len(rl) > 0 {
					n, err = strconv.Atoi(rl)
					if err != nil || n < 1 {
						err = errors.New("invalid RLE format - bad run length")
					}
				}
				rl = ""
				for c := 0; c < n && start+c < mx; c++ {
					result[start+c] = true
				}
				start += n
			case 'b':
				// dead
				n := 1
				if len(rl) > 0 {
					n, err = strconv.Atoi(rl)
					if err != nil || n < 1 {
						err = errors.New("invalid RLE format - bad run length")
					}
				}
				rl = ""
				for c := 0; c < n && start+c < mx; c++ {
					result[start+c] = false
				}
				start += n
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				rl += string(byte(ch))
			default:
				err = errors.New("invalid RLE format")
				return
			}
		}
	}
	return
}
