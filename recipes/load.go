package recipes

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/marrow16/gogol/logic"
	"github.com/marrow16/gogol/patterns"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func Load(filename string) (*Recipe, error) {
	ap, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}
	filename = ap
	f, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("recipe file not found")
		} else {
			return nil, err
		}
	}
	defer func() {
		_ = f.Close()
	}()
	result := &Recipe{}
	if err = json.NewDecoder(f).Decode(result); err != nil {
		return nil, fmt.Errorf("error parsing recipe json: %w", err)
	}
	return result, result.Validate(filename)
}

func (r *Recipe) Validate(filename string) error {
	if gs := r.GridSettings; gs != nil {
		if gs.Width != nil && *gs.Width < 2 {
			return fmt.Errorf("grid width must be greater than 1")
		}
		if gs.Height != nil && *gs.Height < 2 {
			return fmt.Errorf("grid height must be greater than 1")
		}
		if gs.Rule != nil {
			if _, err := logic.NewRuleRle("", *gs.Rule); err != nil {
				return err
			}
		}
	}
	base := filepath.Dir(filename)
	files := make([]*os.File, 0)
	defer func() {
		for _, f := range files {
			_ = f.Close()
		}
	}()
	// check that var names don't collide with pattern names...
	for k := range r.Vars {
		if _, ok := r.Patterns[k]; ok {
			return fmt.Errorf("variable name %q collides with pattern name", k)
		}
	}
	// check that patterns are ok (file exists, name exists, rle is ok, etc.)...
	for n, p := range r.Patterns {
		if _, ok := r.Vars[n]; ok {
			return fmt.Errorf("pattern name %q collides with variable name", n)
		}
		if p != nil {
			switch {
			case p.Name != nil:
				// check that pattern exists in library...
				if patt, ok := patterns.PatternLibrary[*p.Name]; ok {
					p.pattern = &patterns.Pattern{
						Width:  patt.Width,
						Height: patt.Height,
						Cells:  slices.Clone(patt.Cells),
					}
				} else {
					return fmt.Errorf("pattern %q cannot find %q in library", n, *p.Name)
				}
			case p.Filename != nil:
				// check file exists and can be decoded...
				ap, err := resolveEmbeddedPath(base, *p.Filename)
				if err != nil {
					return fmt.Errorf("error resolving path %q: %w", *p.Filename, err)
				}
				p.Filename = &ap
				f, err := os.Open(*p.Filename)
				if err != nil {
					if os.IsNotExist(err) {
						return fmt.Errorf("pattern %q file not found", n)
					}
					return fmt.Errorf("pattern %q opening file %q: %w", n, *p.Filename, err)
				}
				files = append(files, f)
				patt, err := patterns.NewPatternFromRle(f)
				if err != nil {
					return fmt.Errorf("pattern %q error decoding: %w", n, err)
				}
				p.pattern = &patt
			case p.Rle != nil:
				//check that rle can be decoded
				if p.Width == nil {
					return fmt.Errorf("pattern %q requires property 'width'", n)
				} else if *p.Width < 1 {
					return fmt.Errorf("pattern %q bad width %d", n, *p.Width)
				}
				if p.Height == nil {
					return fmt.Errorf("pattern %q requires property 'height'", n)
				} else if *p.Width < 1 {
					return fmt.Errorf("pattern %q bad height %d", n, *p.Height)
				}
				rle := fmt.Sprintf("x = %d, y = %d\n%s!", *p.Width, *p.Height, strings.TrimSuffix(*p.Rle, "!"))
				patt, err := patterns.NewPatternFromRle(strings.NewReader(rle))
				if err != nil {
					return fmt.Errorf("pattern %q error decoding: %w", n, err)
				}
				p.pattern = &patt
			default:
				return fmt.Errorf("pattern %q has insufficient information", n)
			}
		} else {
			return fmt.Errorf("pattern %q is null", n)
		}
	}
	// check that all place properties point at a var or pattern...
	if err := checkDosPlace(r.Do, r); err != nil {
		return err
	}
	return nil
}

func checkDosPlace(dos []Do, recipe *Recipe) error {
	for _, d := range dos {
		if d.Place != nil {
			if _, ok := recipe.Patterns[*d.Place]; !ok {
				if _, ok = recipe.Vars[*d.Place]; !ok {
					return fmt.Errorf("variable or pattern %q does not exist", *d.Place)
				}
			}
		}
		if err := checkDosPlace(d.Do, recipe); err != nil {
			return err
		}
	}
	return nil
}

func resolveEmbeddedPath(base, path string) (string, error) {
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}
	return filepath.Abs(filepath.Join(base, path))
}
