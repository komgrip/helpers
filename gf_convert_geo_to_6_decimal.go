package helpers

import (
	"fmt"
	"strconv"
)

// รับพิกัด latitude หรือ longitude ที่เป็น string แล้วแปลงเป็น string ที่เป็นทศนิมยม 6 ตำแหน่ง
func ConvertGeoTo6Decimal(str string) (string, error) {
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%.6f", f), nil
}
