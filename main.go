package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/google/hilbert"
	"github.com/lucasb-eyer/go-colorful"
	"image"
	"image/png"
	"io"
	"log"
	"os"
)

func ToByte(f float64) byte {
	if f >= 255 {
		return 255
	} else if f <= 0 {
		return 0
	} else {
		return byte(f)
	}
}

func PickN(dx int, dy int) int {
	n := 1

	for n < dx || n < dy {
		n *= 2
	}

	return n
}

func Decode(r io.Reader) (image.Image, error) {
	header := make([]byte, 3+2+2)

	if _, err := io.ReadFull(r, header); err != nil {
		return nil, err
	}

	if !bytes.Equal(header[0:3], []byte("LAB")) {
		return nil, fmt.Errorf("Expected magic bytes: LAB")
	}

	dx := int(binary.LittleEndian.Uint16(header[3:5]))
	dy := int(binary.LittleEndian.Uint16(header[5:7]))

	n := PickN(dx, dy)

	m := &image.RGBA{
		Pix:    make([]uint8, 4*dx*dy),
		Stride: 4 * dx,
		Rect:   image.Rect(0, 0, dx-1, dy-1),
	}

	buf := make([]byte, 4)

	space, err := hilbert.New(n)
	if err != nil {
		return nil, err
	}

	for d := 0; d < n*n; d++ {
		x, y, err := space.Map(d)
		if err != nil {
			return nil, err
		}

		if x >= dx || y >= dy {
			continue
		}

		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, err
		}

		color := colorful.Hcl(
			float64(buf[0])/255*359,
			float64(buf[1])/255*2-1,
			float64(buf[2])/255*2-1,
		)

		a := buf[3]

		m.Pix[y*m.Stride+x*4+0] = ToByte(color.R * float64(a))
		m.Pix[y*m.Stride+x*4+1] = ToByte(color.G * float64(a))
		m.Pix[y*m.Stride+x*4+2] = ToByte(color.B * float64(a))
		m.Pix[y*m.Stride+x*4+3] = byte(a)
	}

	return m, nil
}

func Encode(w io.Writer, m image.Image) error {
	bounds := m.Bounds()

	n := PickN(bounds.Dx(), bounds.Dy())

	header := make([]byte, 3+2+2)

	header[0] = 'L'
	header[1] = 'A'
	header[2] = 'B'

	binary.LittleEndian.PutUint16(header[3:5], uint16(bounds.Dx()))
	binary.LittleEndian.PutUint16(header[5:7], uint16(bounds.Dy()))

	if _, err := w.Write(header); err != nil {
		return err
	}

	buf := make([]byte, 4)

	space, err := hilbert.New(n)
	if err != nil {
		return err
	}

	for d := 0; d < n*n; d++ {
		x, y, err := space.Map(d)
		if err != nil {
			return err
		}

		x += bounds.Min.X
		y += bounds.Min.Y

		if x >= bounds.Max.X || y >= bounds.Max.Y {
			continue
		}

		r, g, b, a := m.At(x, y).RGBA()

		color := colorful.Color{
			float64(r) / float64(a),
			float64(g) / float64(a),
			float64(b) / float64(a),
		}

		h, c, l := color.Hcl()

		buf[0] = byte((h / 360) * 255)
		buf[1] = byte(((c + 1) / 2) * 255)
		buf[2] = byte(((l + 1) / 2) * 255)
		buf[3] = byte(a)

		if _, err := w.Write(buf); err != nil {
			return err
		}
	}

	return nil
}

func Usage() {
	fmt.Printf("Usage: %s [encode | decode]", os.Args[0])
}

func main() {
	if len(os.Args) < 2 {
		Usage()
		return
	}

	switch os.Args[1] {
	case "encode":
		m, _, err := image.Decode(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}

		if err := Encode(os.Stdout, m); err != nil {
			log.Fatal(err)
		}
	case "decode":
		m, err := Decode(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}

		if err := png.Encode(os.Stdout, m); err != nil {
			log.Fatal(err)
		}
	default:
		Usage()
	}
}
