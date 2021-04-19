package main

import (
	"fmt"

	"github.com/gographics/imagick/imagick"
)

func main() {
	fmt.Println("start!!")
	imagick.Initialize()
	defer imagick.Terminate()

	mw1 := imagick.NewMagickWand()
	defer mw1.Destroy()
	err := mw1.ReadImage("neko.png")
	if err != nil {
		panic(err)
	}
	w1 := mw1.GetImageWidth()
	h1 := mw1.GetImageHeight()
	fmt.Println(w1)
	fmt.Println(h1)
}
