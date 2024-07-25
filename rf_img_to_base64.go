package helpers

import (
	"encoding/base64"
	"io/ioutil"
)

func IMGToBase64(imagePath string) (string, error) {
	// อ่านไฟล์จาก path ที่ได้รับมา
	imageBytes, err := ioutil.ReadFile(imagePath)
	if err != nil {
		return "", err
	}

	// แปลงไฟล์รูปภาพเป็น base64
	base64Image := base64.StdEncoding.EncodeToString(imageBytes)

	return base64Image, nil
}
