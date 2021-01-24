package main

import (
	"log"

	"github.com/lehaisonmath6/thumbnaild"
)

func main() {
	thumbnailhelper := thumbnaild.NewThumbailHelper("", "")
	thumbURL, err := thumbnailhelper.GetThumbnailVideo("https://mediacloud.mobilelab.vn/2021-01-14/14_18_13-3c347d70-2058-4879-862b-4d1e23bb7485.mp4")
	if err != nil {
		log.Println("err", err)
	}
	log.Println("thumb", thumbURL)
}
