module github.com/hairglasses-studio/mapitall

go 1.26.1

require (
	github.com/BurntSushi/toml v1.5.0
	github.com/fsnotify/fsnotify v1.8.0
	github.com/godbus/dbus/v5 v5.1.0
	github.com/hairglasses-studio/mapping v0.1.1-0.20260405143738-f2666d440fbd
	github.com/hairglasses-studio/mcpkit v0.3.1-0.20260405143724-9e0ed2c34989
	nhooyr.io/websocket v1.8.17
)

require golang.org/x/sys v0.41.0 // indirect

replace github.com/hairglasses-studio/mapping => ../mapping
