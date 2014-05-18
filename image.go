package main

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dchest/uniuri" // give us random URIs
	"github.com/nfnt/resize"
	"github.com/oliamb/cutter"
)

// Take the site image and save in our server
func GetImage(db DB, imageUrl string) (*Image, error) {

	res, err := http.Get(imageUrl)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	image, err := SaveImage(res.Body)
	if err != nil {
		return nil, err
	}

	return image, nil
}

type dimension struct {
	Width  int
	Height int
}

func SaveImage(file io.Reader) (*Image, error) {

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
	originalImg, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	originalWidht := originalImg.Bounds().Dx()
	originalHeight := originalImg.Bounds().Dy()
	maxSize := ""

	log.Printf("\nPreparing to save de image %dx%d\n", originalWidht, originalHeight)

	// We will save 3 resized images
	sizes := [...]string{"small", "medium", "large"}
	dimensions := [...]dimension{{233, 127}, {358, 195}, {660, 360}}

	// We will save the Thumbnails just if the original image
	// dimension is bigger or equals than the thumbnail dimension
	for index, size := range sizes {
		newWidth := dimensions[index].Width
		newHeight := dimensions[index].Height
		widthRatio := float32(originalWidht) / float32(newWidth)    // Width ratio
		heightRatio := float32(originalHeight) / float32(newHeight) // Height ratio

		// log.Printf("Trying to save img size %d, widthRatio: %f, heightRatio: %f\n", size, widthRatio, heightRatio)

		// We will resize and crop just images bigger than thumbnail in both dimensions
		// If image isn't bigger enought, so we will not take the large image
		if size == "small" || size == "medium" || (widthRatio >= 1 && heightRatio >= 1) {
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

			// If the image is bigger in both directions
			if widthRatio >= 1 && heightRatio >= 1 {
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

			} else {
				// If the image isn't bigger in both directions
				// We will enter here just if its the small/medium size image

				if originalHeight < newHeight {
					// If this image isn't taller enought
					newHeight = originalHeight
				}
				if originalWidht < newWidth {
					// If this image isn't large enought
					newWidth = originalWidht
				}
				// If image still bigger, lets crop it to the right size
				croppedImg, err := cutter.Crop(originalImg, cutter.Config{
					Width:  newWidth,
					Height: newHeight,
					Mode:   cutter.Centered,
				})
				if err != nil {
					return nil, err
				}

				png.Encode(fo, croppedImg)
			}

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
