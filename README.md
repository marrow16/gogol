# gogol

Game Of Life implementation in Go.

## Features

* Easy-to-use terminal UI (TUI)
* Fast rendering and simulation (up to ~60 FPS)
* Step ahead (steps the simulation ahead without rendering)
* Snapshot current grid and revert back to snapshot
* Full control over grid settings
* Standard and custom Life rules
* Built-in pattern library
* Load individual RLE patterns and/or entire libraries
* Pattern preview and metadata viewer
* Pattern placement, rotation and positioning
* Pattern capture from a simulation
  * Interactive pattern editor (cropping and cleanup; metadata editing; save as RLE)
* Grid recipes - JSON files to create initial grid
* Save and load current grid as RLE
* Mouse and keyboard support

## Running

(requires Go 1.26 installed)

TUI (terminal UI):
```
go run ./cmd/tui
```

## Screenshots

![screenshot](./_screenshots/screenshot1.png)
![screenshot](./_screenshots/screenshot2.png)
![screenshot](./_screenshots/screenshot3.png)
![screenshot](./_screenshots/screenshot4.png)
![screenshot](./_screenshots/screenshot5.png)
![screenshot](./_screenshots/screenshot6.png)
![screenshot](./_screenshots/screenshot7.png)
![screenshot](./_screenshots/screenshot8.png)
![screenshot](./_screenshots/screenshot9.png)
![screenshot](./_screenshots/screenshot10.png)

