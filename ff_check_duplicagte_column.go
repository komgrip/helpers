package helpers

import "encoding/json"

// ตรวจสอบcolumn ที่เซ็คใน DB ไว้ว่า unique ไปว่ามีการinsert ซ้ำหรือเปล่า
func CheckDuplicateColumn(err error) bool {
	type GormErr struct {
		Code    string `json:"Code"`
		Message string `json:"Message"`
	}
	byteErr, _ := json.Marshal(err)
	var newError GormErr
	json.Unmarshal((byteErr), &newError)

	return newError.Code == "23505"
}
