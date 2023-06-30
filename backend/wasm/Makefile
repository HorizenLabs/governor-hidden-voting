all: assets/main.wasm

assets/main.wasm: cmd/wasm/*.go
	GOOS=js GOARCH=wasm go build -C cmd/wasm -o ../../assets/main.wasm