#!/bin/sh

image2exif2jsonmap="${HOME}/.cargo/bin/rs-exif2jsonmap.wasm"

export ENV_WASI_PATH="${image2exif2jsonmap}"

inputimg=./input.webp

test -f "${image2exif2jsonmap}" || exec env fname="${image2exif2jsonmap}" sh -c '
	echo wasi byte code "${fname}" missing.
	echo you can find an impl from:
	echo github.com/takanoriyanagitani/rs-exif2jsonmap
	exit 1
'

test -f "${inputimg}" || exec env iname="${inputimg}" sh -c '
	echo input image "${iname}" missing.
	echo you need to find an image with exif to run this demo.
	exit 1
'

echo "--- Version ---"
./cmd/exif2yaml/exif2yaml --version

echo "\n--- Success (using flags) ---"
./cmd/exif2yaml/exif2yaml --wasi-path "${image2exif2jsonmap}" < "${inputimg}" | head -n 4

echo "\n--- Success (using env var) ---"
cat "${inputimg}" |
	./cmd/exif2yaml/exif2yaml |
	head -n 4

echo "\n--- Failure: Wrong WASI path ---"
./cmd/exif2yaml/exif2yaml --wasi-path ./non-existent.wasm < "${inputimg}"

echo "\n--- Failure: Image too small (--image-size-max 10) ---"
./cmd/exif2yaml/exif2yaml --image-size-max 10 < "${inputimg}"

echo "\n--- Failure: WASI too small (--wasi-bytes-max 10) ---"
./cmd/exif2yaml/exif2yaml --wasi-bytes-max 10 < "${inputimg}"
