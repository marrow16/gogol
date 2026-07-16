# GoGoL GUI Help

Note: `⌥` (Mac Option) is `Alt` on Windows

## Statusbar

The statusbar is divided into three sections:

* Left - status (e.g. steps or mode information)
* Middle - current life rule (click or press `⌥L` to change rule)
* Right - control buttons

#### Control buttons
| button                                                | key           | action                    |
|-------------------------------------------------------|---------------|---------------------------|
| <img src="cmd/gui/icons/play.png" width="16">         | `⌥Enter`      | Start/stop the simulation |
| <img src="cmd/gui/icons/step.png" width="16">         | `⌥Space` `⌥→` | Step the simulation       |
| <img src="cmd/gui/icons/skip-forward.png" width="16"> | `⌥Tab`        | Step ahead the simulation |
| <img src="cmd/gui/icons/zoomIn.png" width="16">       | `⌥=`          | Zoom in                   |
| <img src="cmd/gui/icons/zoomOut.png" width="16">      | `⌥-`          | Zoom out                  |
| <img src="cmd/gui/icons/burger.png" width="16">       | `⌥M`          | Menu                      |

When record instrument is enabled (see **Menu > Instrumentation**):

| button                                                 | key          | action                       |
|--------------------------------------------------------|--------------|------------------------------|
| <img src="cmd/gui/icons/backward.png" width="16">      | `⌥←`         | Step backward                |
| <img src="cmd/gui/icons/skip-backward.png" width="16"> | `⌥Backspace` | Skip backward the simulation |

## General keys

| key(s)                | action                          |
|-----------------------|---------------------------------|
| `⌥Enter`              | Start/stop the simulation       |
| `⌥Esc`                | Stop the simulation             |
| `⌥Space` `⌥→`         | Step the simulation             |
| `⌥Tab`                | Step ahead the simulation       |
| `⌥=`                  | Zoom in                         |
| `⌥-`                  | Zoom out                        |
| `⌥B`                  | Toggle cell borders on/off      |
| `⌥C`                  | Clear grid                      |
| `⌥E`                  | Edit mode                       |
| `⌥G`                  | Run grid recipe (when selected) |
| `⌥H`                  | Show heat map (when enabled)    |
| `⌥L`                  | Life rule editor                |
| `⌥M`                  | Menu                            |
| `⌥N`                  | Random noise on grid            |
| `⌥P`                  | Place pattern mode              |
| `⌥R`                  | Randomize grid                  |
| `⌥S`                  | Snapshot grid                   |
| `⌥X`                  | Export grid                     |
| `⌥Z`                  | Undo to snapshot                |
| `⌥,`                  | Decrement life rule permutation |
| `⌥.`                  | Increment life rule permutation |
| `⌥[`                  | Decrease grid width             |
| `⌥]`                  | Increase grid width             |
| `⌥;`                  | Decrease grid height            |
| `⌥'`                  | Increase grid height            |
| `Ctrl+0` ... `Ctrl+8` | Toggle born with                |
| `⌥0` ... `⌥8`         | Toggle survives with            |

## Edit mode keys

| key(s)                                  | action                              |
|-----------------------------------------|-------------------------------------|
| `⌥Esc`                                  | Exit edit mode                      |
| `Space`                                 | Clear cell                          |
| `Shift+Space`                           | Set cell                            |
| `←` `→` `↑` `↓`                         | Move cell cursor                    |
| `Home`                                  | Move cell to beginning of row       |
| `End`                                   | Move cell to end of row             |
| `PgUp`                                  | Move cell to top of grid            |
| `PgDown`                                | Move cell to bottom of grid         |
| `⌥←` `⌥→` `⌥↑` `⌥↓`                     | Draw lines                          |
| `Shift⌥←` `Shift⌥→` `Shift⌥↑` `Shift⌥↓` | Clear lines                         |
| `Shift←` `Shift→` `Shift↑` `Shift↓`     | Mark area                           |
| `Enter`                                 | Capture marked area as pattern      |
| `⌥C`                                    | Clear entire grid                   |
| `⌥F`                                    | Fill marked area with alive cells   |
| `Shift⌥F`                               | Fill marked area with dead cells    |
| `⌥I`                                    | Toggle pattern placement interlaced |
| `⌥P`                                    | Place pattern                       |
| `⌥R`                                    | Pattern placement rotation          |
| `⌥U`                                    | Shift entire grid up                |
| `⌥D`                                    | Shift entire grid down              |
| `⌥,`                                    | Shift entire grid left              |
| `⌥.`                                    | Shift entire grid right             |
| `⌘A` `Ctrl+A`                           | Mark entire grid                    |
| `⌘C` `Ctrl+C`                           | Copy marked area as pattern RLE     |
| `⌘V` `Ctrl+V`                           | Paste pattern RLE                   |
| `⌘X` `Ctrl+X`                           | Cut marked area as pattern RLE      |
| `⌘Z` `Ctrl+Z`                           | Undo                                |
| _character keys_                        | Draw character                      |


