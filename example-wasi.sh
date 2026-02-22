#!/bin/sh

image2exif2jsonmap="${HOME}/.cargo/bin/rs-exif2jsonmap.wasm"

inputimg=./input.webp

e2jwasm=./exif2yaml.wasm
owasm=./opt.wasm

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

test -f "${e2jwasm}" || exec env wname="${e2jwasm}" sh -c '
	echo exif to yaml wasi byte code "${wname}" missing.
	echo you need to built it using ./wasi.sh.
	exit 1
'

run_wa0_on_wa0() {
	cat "${inputimg}" |
		wazero \
			run \
			-mount "${HOME}/.cargo/bin:/guest.d:ro" \
			"${e2jwasm}" \
			-wasi-bytes-max 16777216 \
			-wasi-path /guest.d/rs-exif2jsonmap.wasm |
		bat --language=yaml
}

run_wa0_on_wa0_opt() {
	cat "${inputimg}" |
		wazero \
			run \
			-mount "${HOME}/.cargo/bin:/guest.d:ro" \
			"${owasm}" \
			-wasi-bytes-max 16777216 \
			-wasi-path /guest.d/rs-exif2jsonmap.wasm |
		bat --language=yaml
}

run_wa0_on_wasmtime() {
	cat "${inputimg}" |
		wasmtime \
			run \
			--dir "${HOME}/.cargo/bin::/guest.d" \
			"${e2jwasm}" \
			-wasi-bytes-max 16777216 \
			-wasi-path /guest.d/rs-exif2jsonmap.wasm |
		bat --language=yaml
}

run_wa0_on_wasmtime_opt() {
	cat "${inputimg}" |
		wasmtime \
			run \
			--dir "${HOME}/.cargo/bin::/guest.d" \
			"${owasm}" \
			-wasi-bytes-max 16777216 \
			-wasi-path /guest.d/rs-exif2jsonmap.wasm |
		bat --language=yaml
}

date +"%T.%N"
time run_wa0_on_wasmtime
#time run_wa0_on_wasmtime_opt

date +"%T.%N"
time run_wa0_on_wa0
#time run_wa0_on_wa0_opt
