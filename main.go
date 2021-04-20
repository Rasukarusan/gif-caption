package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

func main() {
	fmt.Println("start!!")
	filename := os.Args[1]
	f, err := os.Open(filename)
	if err != nil {
		log.Fatalf("cannot open file %q: %v", filename, err)
	}
	defer f.Close()

	// GIF分割
	splitGIF(f)

	// 文字を挿入したい画像を読み込む
	file, err := os.Open("nekoko.png")
	if err != nil {
		log.Fatalf("failed to open file: %s", err.Error())
	}
	defer file.Close()

	// 文字を挿入
	addLabel(file)

}

func loadFont() (font *truetype.Font) {
	ttf, err := ioutil.ReadFile("851MkPOP_002.ttf")
	if err != nil {
		log.Fatalf("failed to load font: %s", err.Error())
	}
	ft, err := truetype.Parse(ttf)
	if err != nil {
		log.Fatalf("failed to parse font: %s", err.Error())
	}
	return ft
}

func addLabel(reader io.Reader) {
	img, err := png.Decode(reader)
	if err != nil {
		log.Fatalf("failed to decode image: %s", err.Error())
	}
	dst := image.NewRGBA(img.Bounds())
	draw.Draw(dst, dst.Bounds(), img, image.Point{}, draw.Src)

	col := color.RGBA{125, 184, 236, 1.0}
	opt := truetype.Options{
		Size: 40,
	}
	ft := loadFont()
	face := truetype.NewFace(ft, &opt)

	x, y := 100, 100
	dot := fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(y * 26)}

	d := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(col),
		Face: face,
		Dot:  dot,
	}
	text := "てすと"
	d.DrawString(text)

	newFile, err := os.Create("out.png")
	if err != nil {
		log.Fatalf("failed to create file: %s", err.Error())
	}
	defer newFile.Close()

	b := bufio.NewWriter(newFile)
	if err := png.Encode(b, dst); err != nil {
		log.Fatalf("failed to encode image: %s", err.Error())
	}

}

// @thanks https://stackoverflow.com/questions/33295023/how-to-split-gif-into-images
// Decode reads and analyzes the given reader as a GIF image
func splitGIF(reader io.Reader) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error while decoding: %s", r)
		}
	}()

	gif, err := gif.DecodeAll(reader)

	if err != nil {
		return err
	}

	imgWidth, imgHeight := getGifDimensions(gif)

	overpaintImage := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	draw.Draw(overpaintImage, overpaintImage.Bounds(), gif.Image[0], image.ZP, draw.Src)

	for i, srcImg := range gif.Image {
		draw.Draw(overpaintImage, overpaintImage.Bounds(), srcImg, image.ZP, draw.Over)

		// save current frame "stack". This will overwrite an existing file with that name
		file, err := os.Create(fmt.Sprintf("%s%d%s", "temp", i, ".png"))
		if err != nil {
			return err
		}

		err = png.Encode(file, overpaintImage)
		if err != nil {
			return err
		}

		file.Close()
	}

	return nil
}

func getGifDimensions(gif *gif.GIF) (x, y int) {
	var lowestX int
	var lowestY int
	var highestX int
	var highestY int

	for _, img := range gif.Image {
		if img.Rect.Min.X < lowestX {
			lowestX = img.Rect.Min.X
		}
		if img.Rect.Min.Y < lowestY {
			lowestY = img.Rect.Min.Y
		}
		if img.Rect.Max.X > highestX {
			highestX = img.Rect.Max.X
		}
		if img.Rect.Max.Y > highestY {
			highestY = img.Rect.Max.Y
		}
	}

	return highestX - lowestX, highestY - lowestY
}
