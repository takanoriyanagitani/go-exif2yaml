package img2exif2jsonmap

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	ey "github.com/takanoriyanagitani/go-exif2yaml"
	w0 "github.com/tetratelabs/wazero"
	wa "github.com/tetratelabs/wazero/api"
)

type Instance struct{ wa.Module }

type WasiBytes []byte

func (b WasiBytes) Compile(
	ctx context.Context,
	rtm w0.Runtime,
) (w0.CompiledModule, error) {
	return rtm.CompileModule(ctx, b)
}

func (b WasiBytes) ToConverter(
	ctx context.Context,
	rtm w0.Runtime,
	cfg w0.ModuleConfig,
) (Converter, error) {
	compiled, err := b.Compile(ctx, rtm)
	return Converter{
		CompiledModule: compiled,
		ModuleConfig:   cfg,
	}, err
}

func (i Instance) Close(ctx context.Context) error {
	if nil == i.Module {
		return nil
	}
	return i.Module.Close(ctx)
}

type Config struct{ w0.ModuleConfig }

func (c Config) WithStdin(rdr io.Reader) Config {
	return Config{
		ModuleConfig: c.
			ModuleConfig.
			WithStdin(rdr),
	}
}

func (c Config) WithStderr(wtr io.Writer) Config {
	return Config{
		ModuleConfig: c.
			ModuleConfig.
			WithStderr(wtr),
	}
}

func (c Config) WithStdout(wtr io.Writer) Config {
	return Config{
		ModuleConfig: c.
			ModuleConfig.
			WithStdout(wtr),
	}
}

func (c Config) WithName(name string) Config {
	return Config{
		ModuleConfig: c.
			ModuleConfig.
			WithName(name),
	}
}

type Converter struct {
	w0.CompiledModule
	w0.ModuleConfig
}

func (c Converter) Close(ctx context.Context) error {
	return c.CompiledModule.Close(ctx)
}

func (c Converter) Instantiate(
	ctx context.Context,
	rtm w0.Runtime,
) (Instance, error) {
	ins, err := rtm.InstantiateModule(
		ctx,
		c.CompiledModule,
		c.ModuleConfig,
	)

	return Instance{ins}, err
}

func (c Converter) ToConverter(rtm w0.Runtime) ey.ImageToExifRaw {
	return func(
		ctx context.Context,
		img ey.Image,
	) (ey.RawExifMap, error) {
		var irdr io.Reader = bytes.NewReader(img)
		var obuf bytes.Buffer

		cfg := Config{ModuleConfig: c.ModuleConfig}.
			WithStderr(io.Discard).
			WithStdin(irdr).
			WithStdout(&obuf)
		conv := Converter{
			CompiledModule: c.CompiledModule,
			ModuleConfig:   cfg.ModuleConfig,
		}

		ins, err := conv.Instantiate(
			ctx,
			rtm,
		)

		return obuf.Bytes(), errors.Join(err, ins.Close(ctx))
	}
}

type WazeroConfig struct{ w0.RuntimeConfig }

func (c WazeroConfig) MakeCancelFriendly() WazeroConfig {
	return WazeroConfig{
		RuntimeConfig: c.
			RuntimeConfig.
			WithCloseOnContextDone(true),
	}
}

func (c WazeroConfig) CreateRuntime(ctx context.Context) w0.Runtime {
	return w0.NewRuntimeWithConfig(ctx, c.RuntimeConfig)
}

//nolint:gochecknoglobals
var WazeroConfigDefault WazeroConfig = WazeroConfig{
	RuntimeConfig: w0.NewRuntimeConfig(),
}.MakeCancelFriendly()

type WasiBytesSource func(context.Context) (WasiBytes, error)

func (s WasiBytesSource) ToConverter(
	ctx context.Context,
	rtm w0.Runtime,
	cfg w0.ModuleConfig,
) (Converter, error) {
	wbytes, err := s(ctx)
	if nil != err {
		return Converter{}, err
	}

	return wbytes.ToConverter(
		ctx,
		rtm,
		cfg,
	)
}

type FsSource func(ctx context.Context, wasiPath string) (WasiBytes, error)

func (f FsSource) ToWasiSource(wasiPath string) WasiBytesSource {
	return func(ctx context.Context) (WasiBytes, error) {
		return f(ctx, wasiPath)
	}
}

type WasiBytesMax int64

func (m WasiBytesMax) ToFsSourceOs() FsSource {
	return func(_ context.Context, wasiPath string) (WasiBytes, error) {
		file, err := os.Open(wasiPath) //nolint:gosec // trusts guard by wazero
		if nil != err {
			return nil, fmt.Errorf("%w: %s", err, wasiPath)
		}
		defer file.Close() //nolint:errcheck // ignore close error for read only file

		limited := &io.LimitedReader{
			R: file,
			N: int64(m),
		}

		var buf bytes.Buffer
		_, err = io.Copy(&buf, limited)

		return buf.Bytes(), err
	}
}

const WasiBytesMaxDefault WasiBytesMax = 16777216
