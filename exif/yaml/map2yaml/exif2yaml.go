package exif2yaml

import (
	"context"
	"errors"
	"io"

	ey "github.com/takanoriyanagitani/go-exif2yaml"
	my "github.com/takanoriyanagitani/go-map2yaml"
	m2y "github.com/takanoriyanagitani/go-map2yaml/ser/goccy"
)

type MapToYamlConfig m2y.Config

//nolint:gochecknoglobals
var MapToYamlConfigDefault MapToYamlConfig = MapToYamlConfig(
	m2y.
		ConfigDefault.
		WithIndent(2).
		UseLiteralStyleIfMultiline(true).
		WithFlow(false).
		WithIndentSequence(true).
		UseSingleQuote(false).
		EnableAutoInt(),
)

func (c MapToYamlConfig) ToConverter() ey.ExifToYaml {
	return func(
		ctx context.Context,
		emp ey.ExifMap,
		wtr io.Writer,
	) error {
		var enc m2y.Encoder = m2y.Config(c).ToEncoder(wtr)
		err := enc.Write(
			ctx,
			my.InputMap(emp),
		)
		return errors.Join(err, enc.Close())
	}
}
