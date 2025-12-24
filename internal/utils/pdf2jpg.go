package utils

import (
	"errors"
	"fmt"
	"image/jpeg"
	"os"

	"github.com/gen2brain/go-fitz"
)

func Pdf2Jpg(f string) error {
	// 打开PDF文件
	doc, err := fitz.New(f)
	if err != nil {
		return errors.New("failed to open pdf file, err:" + err.Error())
	}
	defer doc.Close()

	// 遍历所有页码转图片
	for i := 0; i < doc.NumPage(); i++ {
		img, err := doc.Image(i)
		if err != nil {
			return errors.New("failed to render pdf page, err:" + err.Error())
		}
		f, err := os.Create(fmt.Sprintf("page_%d.jpg", i+1))
		if err != nil {
			return errors.New("failed to create file, err:" + err.Error())
		}
		defer f.Close()

		jpeg.Encode(f, img, &jpeg.Options{Quality: 100})
	}
	return nil
}
