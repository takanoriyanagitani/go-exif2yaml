package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	ey "github.com/takanoriyanagitani/go-exif2yaml"
	ej "github.com/takanoriyanagitani/go-exif2yaml/exif/image/exif2json/wasi/wa0"
	my "github.com/takanoriyanagitani/go-exif2yaml/exif/yaml/map2yaml"
	w0 "github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

var version string = "0.0.1"

var (
	ErrWasiPathRequired error = errors.New("wasi path is required (use --wasi-path or ENV_WASI_PATH)")
)

type config struct {
	wasiPath     string
	inputImgMax  int64
	wasiBytesMax int64
	showVersion  bool
}

func (c *config) parse() {
	flag.StringVar(&c.wasiPath, "wasi-path", os.Getenv("ENV_WASI_PATH"), "path to wasi file")
	flag.Int64Var(&c.inputImgMax, "image-size-max", 16777216, "maximum input image size")
	flag.Int64Var(&c.wasiBytesMax, "wasi-bytes-max", 16777216, "maximum wasi bytes")
	flag.BoolVar(&c.showVersion, "version", false, "show version")
	flag.Parse()
}

//nolint:gochecknoglobals
var wcfg ej.WazeroConfig = ej.WazeroConfigDefault

//nolint:gochecknoglobals
var mycfg my.MapToYamlConfig = my.MapToYamlConfigDefault

func sub(ctx context.Context, cfg config) error {
	if cfg.showVersion {
		fmt.Printf("exif2yaml version %s\n", version)
		return nil
	}

	if cfg.wasiPath == "" {
		return ErrWasiPathRequired
	}

	var rtm w0.Runtime = wcfg.CreateRuntime(ctx)
	defer func() {
		err := rtm.Close(ctx)
		if nil != err {
			log.Printf("%v\n", err)
		}
	}()

	wasi_snapshot_preview1.MustInstantiate(ctx, rtm)

	var wbmax ej.WasiBytesMax = ej.WasiBytesMax(cfg.wasiBytesMax)
	var fsrc ej.FsSource = wbmax.ToFsSourceOs()
	var wbs ej.WasiBytesSource = fsrc.ToWasiSource(cfg.wasiPath)

	var mcfg w0.ModuleConfig = w0.NewModuleConfig()

	conv, err := wbs.ToConverter(ctx, rtm, mcfg)
	if nil != err {
		return err
	}
	defer func() {
		err := conv.Close(ctx)
		if nil != err {
			log.Printf("%v\n", err)
		}
	}()

	var i2er ey.ImageToExifRaw = conv.ToConverter(rtm)
	var i2e ey.ImageToExif = i2er.ToConverter()

	var e2y ey.ExifToYaml = mycfg.ToConverter()

	ie2y := ey.ImageExifToYaml{
		ImageToExif: i2e,
		ExifToYaml:  e2y,
	}

	return ie2y.StdinToImageToExifToYamlToStdout(ctx, cfg.inputImgMax)
}

func main() {
	var cfg config
	cfg.parse()

	err := sub(context.Background(), cfg)
	if nil != err {
		log.Printf("%v\n", err)
		os.Exit(1)
	}
}
