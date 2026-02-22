package exif2yaml

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
)

type ExifBytes []byte

type Image []byte

type ExifMap map[string]any

type ParseExif func(context.Context, ExifBytes) (ExifMap, error)

type ImageToExif func(context.Context, Image) (ExifMap, error)

type RawExifMap []byte

func (r RawExifMap) ParseJSON() (ExifMap, error) {
	out := map[string]any{}
	err := json.Unmarshal(r, &out)
	return out, err
}

type ImageToExifRaw func(context.Context, Image) (RawExifMap, error)

func (i ImageToExifRaw) ToConverter() ImageToExif {
	return func(ctx context.Context, img Image) (ExifMap, error) {
		raw, err := i(ctx, img)
		if nil != err {
			return nil, err
		}

		return raw.ParseJSON()
	}
}

type ExifToYaml func(context.Context, ExifMap, io.Writer) error

type ImageExifToYaml struct {
	ImageToExif
	ExifToYaml
}

func (i ImageExifToYaml) ImageToExifToYaml(
	ctx context.Context,
	img Image,
	wtr io.Writer,
) error {
	emap, err := i.ImageToExif(ctx, img)
	if nil != err {
		return err
	}

	return i.ExifToYaml(ctx, emap, wtr)
}

func (i ImageExifToYaml) ReaderToImageToExifToYamlToWriter(
	ctx context.Context,
	imgMax int64,
	rdr io.Reader,
	wtr io.Writer,
) error {
	var buf bytes.Buffer
	limited := &io.LimitedReader{
		R: rdr,
		N: imgMax,
	}
	_, err := io.Copy(&buf, limited)
	return errors.Join(
		err,
		i.ImageToExifToYaml(
			ctx,
			buf.Bytes(),
			wtr,
		),
	)
}

func (i ImageExifToYaml) StdinToImageToExifToYamlToStdout(
	ctx context.Context,
	imgMax int64,
) error {
	var bwtr *bufio.Writer = bufio.NewWriter(os.Stdout)
	err := i.ReaderToImageToExifToYamlToWriter(
		ctx,
		imgMax,
		os.Stdin,
		bwtr,
	)
	return errors.Join(err, bwtr.Flush())
}
