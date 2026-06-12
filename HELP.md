# GoGoL TUI Help

## General

The terminal title shows the simulation state - with step count and current life rule.

When in simulation, use the following keys:

| key         | action                     |
|-------------|----------------------------|
| `enter`     | start/stop the simulation  |
| `space`     | step the simulation        |
| `tab`       | step ahead the simulation  |
| `home`      | randomize the grid         |
| `ctrl+s`    | settings                   |
| `ctrl+k`    | capture mode               |
| `ctrl+g`    | grid recipes               |
| `ctrl+o`    | snapshot current grid      |
| `backspace` | restore to last snapshot   |
| `ctrl+x`    | export current grid as RLE |
| `esc`       | quit (with save)           |
| `ctrl+c`    | quit (no save)             |

Note: running, stepping and step ahead will only continue when there are grid state changes.

---

## Settings dialog

Press `ctrl+s` to open the Settings dialog.

### Grid tab

Edit the settings for the grid...

* **Clear** press to clear the current grid.
* **Randomize** press to fill grid with random density.
* **Randomization %** sets the random density.
* **Step delay (ms)** is the delay, in milliseconds, between steps when running simulation.
* **Step ahead size** is the number of steps to run when stepping ahead (use `tab` key during simulation).


* **Snapshot before** determines whether a snapshot of the current grid is taken before stepping ahead.
* **Wrapping mode** sets the wrapping mode for the grid.
* **Boundary mode** sets whether cells at boundary (off edge) are considered dead or alive (when not wrapping).
* **Height** & **Width** adjust the desired size of the grid
  * **Resize** press to resize the grid to the desired size.
  * **Fit Screen** press to resize the grid to fit in the current terminal screen size.


* **Foreground** & **Background**<br>
    Use the **R**, **G** & **B** settings to adjust the color.

### Rule tab

Change the current life rule...

* **Name** select the rule name from currently available rules.  Use up/down arrows to navigate; enter to show dropdown; or a letter to search.
* **Rule** displays the born/survives for the currently selected rule.  Also enter an entirely different rule setting.
* **Permutation** displays the permutation of the currently selected rule.  Also enter a number or use up/down keys to adjust.

Note: The permutation index is calculated by treating the rule's birth and survival conditions as an 18-bit binary number. Bits 9-17 represent birth counts (B0-B8) and bits 0-8 represent survival counts (S0-S8). Each enabled condition contributes 2ⁿ to the index, producing a unique integer representation for every possible Life rule.

When an unknown rule or permutation is entered, further entries will appear - to allow saving the rule as a named rule...
* **As name** enter the name you wish to save the rule as
* **Save** press to save the new named rule, 

### Patterns tab

Select a pattern from the current pattern library and place it on the grid...

* **Name** select the name of a pattern from the currently loaded library. Use up/down arrows to navigate; enter to show dropdown; or a letter to search.
* Preview area shows a preview of the selected pattern.
  * press `ctrl+k` to switch between preview and pattern metadata.
* **At** are the coordinates to place the pattern on the grid
  * **Y** & **X** are the positions - enter a number or use up/down keys to adjust.
  * **Rotate** is whether the pattern should be rotated
  * **Place** press to place the pattern on the grid at the selected coordinates.

### Load tab

Load individual RLE file or an entire library (from a directory)...
* **From** enter the path to a file or directory.  Or paste a path (Mac users can press `ctrl+f` to open using finder).
* **Load** press to load the file or directory into the current pattern library.


* **Clear** to clear all loaded patterns from the current pattern library.

### Export/Import tab

Export the current grid to RLE file or import current grid from RLE file...

* **Export Grid** to save the current grid to RLE file.


* **Import from**  enter the filename of an RLE to load the grid from.
* **Resize** determines whether to resize the current grid to match the RLE dimensions.
* **Import Grid** press to import the selected RLE file.
---

## Grid Recipes dialog

Press `ctrl+g` to open the Grid Recipes dialog.

For more information about Grid Recipes - see [Recipes README](https://github.com/marrow16/gogol/blob/main/recipes/README.md)

### Select tab

Select and run Grid Recipes...

* **Recipe** select the desired recipe. Use up/down arrow to scroll; enter to show dropdown.
* Preview area shows the recipe JSON or preview.
* **Run** press to run the selected recipe
* **Save rle** press to save the selected recipe as an RLE (after running)

### Load tab

Load Grid Recipes from JSON files...

* **From** enter the filename of the recipe JSON (Mac users can press `ctrl+f` to open using finder).
* **Load** press to load the entered recipe JSON file.

---

## Capturing & Capture dialog

Press `ctrl+k` to start capturing.

Once capturing has started, move the cursor to the beginning area of the grid you wish to capture - then press `space` or `enter`.

Once the beginning has been marked, use the cursor to locate the end of the area to be captured - then press `space` or `enter`.

Having marked the area to be captured, the Capture dialog will appear.

### Details tab

Select the filename to save pattern to and modify the RLE metadata...

* **File** enter the filename to save pattern as (a `.rle` extension will be added automatically).
* **Name** is the name of the pattern (for RLE metadata).
* **Comments** enter any comments (RLE metadata) for the pattern.
* **Add pattern** determines whether the pattern will also be added to the current pattern library.
* **Save** press to save the pattern as RLE

### Modify tab

Make final adjustments to the captured pattern...

* Preview area - use the cursor to move around the preview
  * `space`/`backspace` to clear a cell
  * `ctrl+space` to set a cell
* **Crop** use the **Top**, **Left**, **Bottom** and **Right** values to crop the pattern.<br>
  As these values are adjusted the preview is updated (but the crop isn't applied until save).