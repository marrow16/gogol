# Grid Recipes

A Grid Recipe is a JSON document that describes how to construct an initial Game of Life grid.

Recipes are executed on demand and are not preloaded or cached. A recipe can be edited in any text editor and re-run immediately without restarting GoGoL.

Press `ctrl+g` in GoGoL to show the Grid Recipes dialog.  That dialog allows you to load recipes; execute a recipe; save a recipe as .rle

## JSON Structure

The overall structure is...

```json
{
  "name": "Example",
  "grid": { ... },
  "vars": { ... },
  "patterns": { ... },
  "do": [ ... ] }
```

All unknown JSON properties are ignored.

---

### `grid` property

The optional `grid` property configures the overall grid (if this property is not specified, then the current grid settings are used)

```json
{
  "grid": {
    "height": 100,
    "width": 100,
    "rule": "B3/S23",
    "wrap_mode": "toroidal",
    "boundary_mode": "dead"
  }
}
```

All properties are optional - properties omitted (or null) default to the current grid setting.

---

### `vars` property

Variables store numeric values or binary patterns.

```json
{
  "vars": {
    "my-pattern": "0b000101",
    "my-num": 10
  }
}
```

When using a string to denote the initial value...
* `"0b101"` (binary notation) denotes 5 and a specific width (when used as a pattern) of 3 cells
* `"0b000101"` (binary notation) denotes 5 and a specific width (when used as a pattern) of 7 cells
* `"0o07"` (octal notation) denotes 7 and a specific width (when used as a pattern) of 6 cells
* `"0x0E"` (hexadecimal notation) denotes 14 and a specific width (when used as a pattern) of 8 cells
* `"127"` (decimal notation) denotes 127 and the width (when used as a pattern) will be the number of bits to represent that value

Also the string can be prefixed with various flags...
* `"alt:"`|`"alternate:"`<br>
    alternate cells on each successive placement
* `"rot:"`|`"rotate:"`<br>
    rotate cells on each successive placement according to Y position
* `"fw:"`|`"fillwidth:"`|`"fill-width:"`<br>
    the pattern is built to fit the grid width
* `"fh:"`|`"fillheight:"`|`"fill-height:"`<br>
  the pattern is built to fit the grid height

A variable can be placed on the grid (as the derived binary pattern) by referencing its name in a `place` property, e.g.
```json
{
  "do": [
    {
      "place": "my-pattern"
    }
  ]
}
```

---

### `patterns` property

Patterns can be: 
* External rle files
* Inline rle definitions
* References to a name in the current pattern library

##### External rle file
```json
{
  "patterns": {
    "beacon": {
      "filename": "beacon.rle"
    }
  }
}
```
Note: relative filenames are relative from the recipe json file.

The recipe will error if the file cannot be found (or cannot be decoded).

##### Inline rle definition
```json
{
  "patterns": {
    "block": {
      "rle": "2o$2o!",
      "width": 2,
      "height": 2
    }
  }
}
```
Note: the `width` and `height` properties are mandatory when `rle` property is specified.

The recipe will error if the `rle` cannot be decoded.

##### Reference to pattern library
```json
{
  "patterns": {
    "glider": {
      "name": "Glider"
    }
  }
}
```
or abbreviated form...
```json
{
  "patterns": {
    "glider": "Glider"
  }
}
```

The recipe will error if the name cannot be found in the current pattern library.

---

### `do` property

Contains the actual instructions for placing items on the grid.

```json
{
  "do": [
    {
      "place": "name",
      "at": { ... },
      "move": { ... },
      "repeat": 1,
      "rotate": 1,
      "do": [ ... ],
      "var_operations": [ ... ]
  },
    ...
  ]
}
```
All properties are optional.

##### `place` property

The name of the variable or pattern to place (referenced from the `vars`/`patterns` in the recipe).

If the name does not resolve to a variable or pattern, the recipe will error.

If the `place` is omitted or null - nothing happens (but nested `do` instructions are carried out).

##### `at` property

Defines an absolute placement position on the grid, e.g.
```json
"at": {
  "x:": 0,
  "y": 0
}
```
Both the `x` and `y` properties are optional - if not specified, the current position is used.

##### `move` property
Defines a relative position to the current position, e.g.
```json
"move": {
  "x:": 0,
  "y": 0,
  "when": "before|after"
}
```
All properties are optional.

If the `x`/`y` properties are omitted or null then the current position for that part is unaffected.

The `x`/`y` properties are relative to the current position, so can be positive or negative.  They can also be specified as a string - one of the following:
* `"lw"` is the width of the last placed pattern/var
* `"lh"` is the height of the last placed pattern/var
* `"-lw"` is the negative width of the last placed pattern/var
* `"-lh"` is the negative height of the last placed pattern/var

Using the above tokens, the dimension can also be adjusted e.g.
* `"lw+5"` is the last width plus 5
* `"lh++"` is the last width plus 1

The `when` property controls when the move happens - i.e. before or after the placement.  If omitted or null it assumes before.

##### `repeat` property

Is the number of repeats (after the initial placement).

This can also be specified as a string - one of the following:
* `"gh"`|`"gridheight"`|`"grid-height"`<br> 
    repeats to fill the remaining grid height
* `"gw"`|`"gridwidth"`|`"grid-width"`<br>
  repeats to fill the remaining grid width

##### `rotate` property

Is the number of 90 degree rotations for the placed pattern.  Values greater than 3 are taken as modulus 4.

##### `do` property

Contains nested instructions.

##### `var_operations` property

Are operations to perform on the currently placed var (and only relevant when placing a var).

```json
{
  "place": "my-var",
  ...
  "var_operations": [
    {
      "when": "before|after", // defaults to after
      "shift_left": 1,
      "shift_right": 1,
      "rotate-left": 1,
      "rotate-right": 1,
      "increment": 1,
      "decrement": 1
    },
    ...
  ]
}

```

---

## Examples

### Fill grid with chequered cells
```json
{
  "name": "Chequered",
  "grid": {
    "wrap_mode": "toroidal",
    "boundary_mode": "dead"
  },
  "vars": {
    "pattern": "fill-width:rotate:0b01"
  },
  "do": [
    {
      "at": {"x": 0, "y": 0},
      "do": [
        {
          "place": "pattern",
          "repeat": "gh",
          "move": {"y": 1, "when": "after"}
        }
      ]
    }
  ]
}
```