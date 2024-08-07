package helpers

import "gorm.io/gorm"

type Identifiable interface {
	GetID() int
	TableName() string
}

//สร้างmap key value ของmodelที่เป็นmaster data โดยต้องไปกำหนด receiver function ตาม interface Identifiable มาก่อน
//
//เช่น 
//func (b Responsible) GetID() int {
// 	return b.ID
// }
//
//func (b Responsible) TableName() string {
// 	return "responsible"
// }
func MakeMapMaster[T Identifiable](db *gorm.DB, models []T) (map[int]T, error) {
	err := db.Find(&models).Error
	if err != nil {
		return nil, err
	}
	result := make(map[int]T)
	for _, data := range models {
		result[data.GetID()] = data
	}
	result[0] = *new(T) //กรณีค่าIDเป็น 0
	return result, nil
}
