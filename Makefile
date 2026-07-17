.PHONY: build
build: build-mac build-windows

.PHONY: build-mac
build-mac: build-mac-amd build-mac-arm

.PHONY: build-mac-amd
build-mac-amd:
	$(info Building mac-amd64)
	gogio -target macos -arch amd64 -appid com.github.marrow16.gogol -icon gogol-neon-icon.png -o _builds/gui/mac/amd64/GoGoL.app ./cmd/gui
	rm -rf _builds/gui/mac/amd64/dmg
	mkdir -p _builds/gui/mac/amd64/dmg
	cp -R _builds/gui/mac/amd64/GoGoL.app _builds/gui/mac/amd64/dmg/
	ln -sf /Applications _builds/gui/mac/amd64/dmg/Applications
	hdiutil create \
		-volname "GoGoL" \
		-srcfolder _builds/gui/mac/amd64/dmg \
		-ov \
		-format UDZO \
		-imagekey zlib-level=9 \
		_builds/gui/mac/amd64/GoGoL-amd64.dmg
	rm -rf _builds/gui/mac/amd64/dmg

.PHONY: build-mac-arm
build-mac-arm:
	$(info Building mac-arm64)
	gogio -target macos -arch arm64 -appid com.github.marrow16.gogol -icon gogol-neon-icon.png -o _builds/gui/mac/arm64/GoGoL.app ./cmd/gui
	rm -rf _builds/gui/mac/arm64/dmg
	mkdir -p _builds/gui/mac/arm64/dmg
	cp -R _builds/gui/mac/arm64/GoGoL.app _builds/gui/mac/arm64/dmg/
	ln -sf /Applications _builds/gui/mac/arm64/dmg/Applications
	hdiutil create \
		-volname "GoGoL" \
		-srcfolder _builds/gui/mac/arm64/dmg \
		-ov \
		-format UDZO \
		-imagekey zlib-level=9 \
		_builds/gui/mac/arm64/GoGoL-arm64.dmg
	rm -rf _builds/gui/mac/arm64/dmg

.PHONY: build-windows
build-windows: build-windows-amd build-windows-arm

.PHONY: build-windows-amd
build-windows-amd:
	$(info Building windows-amd64)
	gogio -target windows -arch amd64 -icon gogol-neon-icon.png -o _builds/gui/windows/amd64/GoGoL.exe ./cmd/gui

.PHONY: build-windows-arm
build-windows-arm:
	$(info Building windows-arm64)
	gogio -target windows -arch arm64 -icon gogol-neon-icon.png -o _builds/gui/windows/arm64/GoGoL.exe ./cmd/gui
