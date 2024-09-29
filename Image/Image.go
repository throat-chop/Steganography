package Image

import (
	"encoding/binary"
	"fmt"
	_ "golang.org/x/image/bmp"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"os"
	"strconv"
)

type Img struct {
	Data           *image.RGBA64
	HidingCapacity uint32
	Type           string
}

func NewImage(filepath string) (*Img, error) {
	file, err := os.Open(filepath)
	defer file.Close()
	if err != nil {
		return nil, err
	}

	decoded, xtn, err := image.Decode(file)

	if err != nil {

		return nil, err
	}

	decodedRGBA64 := getRGBA64(decoded)

	capacity := uint32((decodedRGBA64.Bounds().Dx() * decodedRGBA64.Bounds().Dy() * 3) / (8 / 8))

	return &Img{
		Data:           decodedRGBA64,
		HidingCapacity: capacity,
		Type:           xtn,
	}, nil
}

func (wf *Img) ImgCheck(data uint32) error {

	abs := int32(data)

	if abs < 0 {
		abs *= -1
	}

	if wf.HidingCapacity <= uint32(abs) {
		return fmt.Errorf("max hiding capacity: " + strconv.Itoa(int(wf.HidingCapacity)) + " out of range for given data")
	}
	return nil
}

func getRGBA64(img image.Image) *image.RGBA64 {
	if rgba64Img, ok := img.(*image.RGBA64); ok {
		return rgba64Img
	}

	bounds := img.Bounds()
	nrgbaImg := image.NewRGBA64(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := color.RGBA64Model.Convert(img.At(x, y)).(color.RGBA64)
			nrgbaImg.SetRGBA64(x, y, c)
		}
	}
	return nrgbaImg
}

func (wf *Img) Hide(data []byte, dir string) error {

	bounds := wf.Data.Bounds()

	if len(data)%3 != 0 {
		data = append(data, make([]byte, 3-(len(data)%3))...)
	}

stop:
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {

			if len(data) < 1 {
				break stop
			}

			originalColor := wf.Data.RGBA64At(x, y)
			newcolor := splitAndHide(data[:3], originalColor)
			wf.Data.SetRGBA64(x, y, newcolor)

			data = data[3:]

		}
	}

	hidden, err := os.Create(dir)
	defer hidden.Close()
	if err != nil {
		return err
	}

	encoder := png.Encoder{CompressionLevel: png.BestCompression}

	err = encoder.Encode(hidden, wf.Data)
	if err != nil {
		return err
	}
	return nil
}

func splitAndHide(byteValue []byte, og color.RGBA64) color.RGBA64 {

	arr := []uint16{og.R, og.G, og.B, og.A}

	for i := 0; i < len(arr)-1; i++ {
		arr[i] = (arr[i] & 0xff00) | uint16(byteValue[i])
	}

	return color.RGBA64{
		R: arr[0],
		G: arr[1],
		B: arr[2],
		A: arr[3],
	}
}

func (wf *Img) Extract() []byte {

	bounds := wf.Data.Bounds()

	m := uint32(0)
	absSize := wf.HidingCapacity
	var data []byte

stop:
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {

			if len(data) >= 4 {
				size := int32(binary.BigEndian.Uint32(data[:4]))
				if size < -1 {
					absSize = uint32(size * -1)
				} else {
					absSize = uint32(size)
				}
			} else if m*3 == absSize {
				break stop
			}

			originalColor := wf.Data.RGBA64At(x, y)
			extractedPart := reverseAndCombine(originalColor)

			data = append(data, extractedPart...)
			m++

		}
	}
	return data[:absSize]
}

func reverseAndCombine(og color.RGBA64) []byte {

	arr := []uint16{og.R, og.G, og.B}

	byteValue := make([]byte, 3)

	for i := 0; i < len(byteValue); i++ {
		byteValue[i] = uint8(arr[i] % 256)
	}

	return byteValue
}
