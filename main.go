package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

const prefixTempFile = "temp"

func main() {
	fmt.Println("start!!")

	defer removeTempFile()

	// 対象のGIFを読み込む
	filename := os.Args[1]
	f, err := os.Open(filename)
	if err != nil {
		log.Fatalf("cannot open file %q: %v", filename, err)
	}
	defer f.Close()

	// GIF分割
	names, err := splitGif(f)
	if err != nil {
		log.Fatalf(err.Error())
	}

	// 最初のフレームに文字を挿入
	for i := 0; i < 5; i++ {
		f1, err := os.Open(names[i])
		if err != nil {
			log.Fatalf("failed to open file: %s", err.Error())
		}
		defer f1.Close()
		addLabel(f1, "START")
	}
	makeGif(names)
}

func removeTempFile() {
	files, err := filepath.Glob(prefixTempFile + "*")
	if err != nil {
		panic(err)
	}
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			panic(err)
		}
	}
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

func addLabel(file *os.File, text string) {
	img, err := png.Decode(file)
	if err != nil {
		log.Fatalf("failed to decode image: %s", err.Error())
	}
	dst := image.NewRGBA(img.Bounds())
	draw.Draw(dst, dst.Bounds(), img, image.Point{}, draw.Src)

	// col := color.RGBA{255, 255, 255, 1.0}
	opt := truetype.Options{
		Size: 40,
	}
	ft := loadFont()
	face := truetype.NewFace(ft, &opt)

	x, y := 100, 100
	dot := fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(y * 26)}

	d := &font.Drawer{
		Dst:  dst,
		Src:  image.White,
		Face: face,
		Dot:  dot,
	}
	d.DrawString(text)

	newFile, err := os.Create(file.Name())
	if err != nil {
		log.Fatalf("failed to create file: %s", err.Error())
	}
	defer newFile.Close()
	b := bufio.NewWriter(newFile)
	if err := png.Encode(b, dst); err != nil {
		log.Fatalf("failed to encode image: %s", err.Error())
	}
	b.Flush()
}

// @thanks https://stackoverflow.com/questions/33295023/how-to-split-gif-into-images
// Decode reads and analyzes the given reader as a GIF image
func splitGif(reader io.Reader) (names []string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error while decoding: %s", r)
		}
	}()

	gif, err := gif.DecodeAll(reader)

	if err != nil {
		return []string{""}, err
	}

	imgWidth, imgHeight := getGifDimensions(gif)

	overpaintImage := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	draw.Draw(overpaintImage, overpaintImage.Bounds(), gif.Image[0], image.ZP, draw.Src)

	// ns := make([]string, len(gif.Image))
	var ns []string
	for i, srcImg := range gif.Image {
		draw.Draw(overpaintImage, overpaintImage.Bounds(), srcImg, image.ZP, draw.Over)

		// save current frame "stack". This will overwrite an existing file with that name
		file, err := os.Create(fmt.Sprintf("%s%d%s", prefixTempFile, i, ".png"))
		if err != nil {
			return []string{""}, err
		}

		err = png.Encode(file, overpaintImage)
		if err != nil {
			return []string{""}, err
		}

		ns = append(ns, file.Name())
		file.Close()
	}

	return ns, nil
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

func makeGif(names []string) {
	outGif := &gif.GIF{}
	for _, name := range names {
		fmt.Println(name)
		f, err := os.Open(name)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		png, _, err := image.Decode(f)
		if err != nil {
			panic(err)
		}
		palettedImage := image.NewPaletted(png.Bounds(), palette.Plan9)
		draw.Draw(palettedImage, palettedImage.Rect, png, png.Bounds().Min, draw.Over)
		outGif.Image = append(outGif.Image, palettedImage)
		outGif.Delay = append(outGif.Delay, 0)
	}
	f, _ := os.OpenFile("out.gif", os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()
	gif.EncodeAll(f, outGif)
}
