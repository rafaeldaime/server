package main

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dchest/uniuri" // give us random URIs
	"github.com/nfnt/resize"
	"github.com/oliamb/cutter"
)

type dimension struct {
	Width  int
	Height int
}

// Take the site image and save in our server
func GetImage(db DB, imageUrl string) (*Image, error) {

	log.Printf("Saving the image %s\n", imageUrl)

	res, err := http.Get(imageUrl)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	imageId := uniuri.NewLen(20)
	i, err := db.Get(Image{}, imageId)
	for err == nil && i != nil {
		// Shit, we generated an existing ImageId!
		// Aren't we so lucky?!
		imageId := uniuri.NewLen(20)
		i, err = db.Get(Image{}, imageId)
	}

	if err != nil {
		return nil, err
	}

	// Lets decode our image to work with it
	originalImg, _, err := image.Decode(res.Body)
	if err != nil {
		return nil, err
	}

	originalWidht := originalImg.Bounds().Dx()
	originalHeight := originalImg.Bounds().Dy()
	maxSize := ""

	log.Printf("\nPreparing to save de image %dx%d\n\n", originalWidht, originalHeight)

	// We will save 3 resized images
	sizes := [...]string{"small", "medium", "large"}
	// dimensions = {{233, 127}, {312, 170}, {661, 360}}
	dimensions := [...]dimension{{233, 127}, {312, 170}, {661, 360}}

	// We will save the Thumbnails just if the original image
	// dimension is bigger or equals than the thumbnail dimension
	for index, size := range sizes {
		newWidth := dimensions[index].Width
		newHeight := dimensions[index].Height
		widthRatio := float32(originalWidht) / float32(newWidth)    // Width ratio
		heightRatio := float32(originalHeight) / float32(newHeight) // Height ratio
		// Neither width or height can be smaller than the thumbnail

		log.Printf("\nDEBUG: %dx%d = %dx%d\n\n", newWidth, newHeight, widthRatio, heightRatio)

		// We will resize and crop just images bigger than thumbnail in both dimensions
		if widthRatio >= 1 && heightRatio >= 1 {
			log.Println("FIADAPUTA")
			maxSize = size
			// Below the values to the image be resized after be cropped
			// If resized width or height == 0, so the proportion will be conserved
			resizedWidth := 0  // Width to image be resized
			resizedHeight := 0 // Height to image be resized
			if widthRatio >= heightRatio {
				// We will resize based on height and crop it after resized
				resizedHeight = newHeight
			} else {
				// We will resize besed on height and crop it after resized
				resizedWidth = newWidth
			}

			fo, err := os.Create("public/img/" + imageId + "-" + size + ".png")
			if err != nil {
				return nil, err
			}

			// close fo on exit and check for its returned error
			defer func() {
				if err := fo.Close(); err != nil {
					panic(err)
				}
			}()

			// Downscales an image preserving its aspect ratio to the maximum dimensions
			resizedImage := resize.Resize(uint(resizedWidth), uint(resizedHeight), originalImg, resize.Lanczos3)

			// If image still bigger, lets crop it to the right size
			croppedImg, err := cutter.Crop(resizedImage, cutter.Config{
				Width:  newWidth,
				Height: newHeight,
				Mode:   cutter.Centered,
			})
			if err != nil {
				return nil, err
			}

			png.Encode(fo, croppedImg)

			log.Printf("Saving the image size %s\n", size)
		}
	}

	if maxSize == "" {
		// There is no image saved, cause the original image is so small
		return nil, nil
	}

	image := &Image{
		ImageId:  imageId,
		MaxSize:  maxSize,
		Creation: time.Now(),
		Deleted:  false,
	}

	err = db.Insert(image)
	if err != nil {
		return nil, err
	}

	return image, nil
}
