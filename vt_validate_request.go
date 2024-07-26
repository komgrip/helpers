package helpers

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	emailValidate "github.com/go-playground/validator/v10"

	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"gopkg.in/validator.v2"
	"gorm.io/gorm"
)

type fail struct {
	Status bool        `json:"status" extensions:"x-order=0"`
	Code   int         `json:"code" extensions:"x-order=1"`
	Err    errorFormat `json:"error" extensions:"x-order=2"`
}

type errorFormat struct {
	Message string      `json:"message" extensions:"x-order=0"`
	Field   interface{} `json:"field" extensions:"x-order=1"`
}

func ValidateResponse(msg map[string]string) fail {
	var request fail
	request.Code = 422
	request.Err.Message = "validate error"
	request.Err.Field = msg
	request.Status = false
	return request
}

func initValidator(c *gin.Context, db *gorm.DB) *validator.Validator {
	validate := validator.NewValidator()
	validate.SetValidationFunc("acceptlist", acceptList)
	validate.SetValidationFunc("date", dateFormat)
	validate.SetValidationFunc("tel", validateTel)
	validate.SetValidationFunc("thaiid", thaiNationalIDValidator)
	validate.SetValidationFunc("email", validateEmail)

	if db != nil {
		validate.SetValidationFunc("unique", validateUniqueValue(c, db))
	}
	return validate
}

// initErrorMessages initializes and returns a nested map containing error messages
// for different validation rules and fields. The outer map's keys represent
// validation errors, while the inner map's keys represent specific field names.
// The "default" key is used for fields that do not have specific error messages.

func initErrorMessages() map[string]map[string]string {
	return map[string]map[string]string{
		"less than min": {
			"tel":      "โปรดระบุหมายเลขโทรศัพท์ให้ครบ 10 หลัก",
			"password": "โปรดระบุรหัสผ่านไม่ต่ำกว่า 12 ตัวอักษร",
			"default":  "ค่าไม่สามารถติดลบได้",
		},
		"zero value":                           {"default": "โปรดระบุ"},
		"regular expression mismatch":          {"default": "รูปแบบอีเมลไม่ถูกต้อง"},
		"incorrect":                            {"default": "ข้อมูลไม่ถูกต้อง"},
		"companyEstablishment value not equal": {"default": "ข้อมูลทุนจัดตั้งที่นำเข้ากับข้อมูลหน้าตั้งค่าไม่เท่ากัน"},
		"mismatch regis":                       {"default": "รหัสผ่านไม่ตรงกัน"},
		"mismatch":                             {"old_password": "รหัสผ่านไม่ถูกต้อง", "default": "รหัสผ่านไม่ตรงกัน"},
		"mismatch with old password":           {"default": "ยืนยันรหัสผ่านไม่ถูกต้อง"},
		"already used":                         {"default": "ข้อมูลนี้ถูกใช้ในระบบแล้ว"},
		"not found":                            {"default": "ไม่พบข้อมูล"},
		"inactive":                             {"default": "ข้อมูลไม่ได้ใช้งาน"},
		"not verify":                           {"default": "ข้อมูลไม่ได้ยืนยัน"},
		"not equal":                            {"default": "ข้อมูลไม่ครบถ้วน"},
		"duplicate":                            {"default": "ข้อมูลนี้ถูกใช้แล้ว"},
		"have space":                           {"default": "โปรดระบุข้อมูลที่ไม่มีช่องว่าง"},
		"password must contain at least one uppercase letter":  {"default": "โปรดระบุตัวพิมพ์ใหญ่อย่างน้อย 1 ตัว"},
		"password must contain at least one lowercase letter":  {"default": "โปรดระบุตัวพิมพ์เล็กอย่างน้อย 1 ตัว"},
		"password must contain at least one digit letter":      {"default": "โปรดระบุตัวเลข 0-9 อย่างน้อย 1 ตัว"},
		"password must contain at least one special character": {"default": "โปรดระบุสัญลักษณ์อย่างน้อย 1 ตัว"},
		"duplicate input":                  {"default": "Code หรือ Name ซ้ำกัน"},
		"companyEstablishment value error": {"default": "รหัสทุนจัดตั้งบริษัทไม่ถูกต้อง"},
		"invalid email address":            {"default": "รูปแบบอีเมลไม่ถูกต้อง"},
		"greater than max": {
			"code":             "โปรดระบุข้อมูลไม่เกิน 100 ตัวอักษร",
			"peak_code":        "โปรดระบุข้อมูลไม่เกิน 100 ตัวอักษร",
			"express_code":     "โปรดระบุข้อมูลไม่เกิน 100 ตัวอักษร",
			"description":      "โปรดระบุข้อมูลไม่เกิน 125 ตัวอักษร",
			"discount_percent": "โปรดระบุข้อมูลเปอร์เซ็นต์ไม่เกิน 100 เปอร์เซ็นต์",
			"name":             "โปรดระบุข้อมูลไม่เกิน 255 ตัวอักษร",
			"detail":           "โปรดระบุข้อมูลไม่เกิน 255 ตัวอักษร",
			"tel":              "โปรดระบุหมายเลขโทรศัพท์ไม่เกิน 10 หลัก",
			"round":            "โปรดระบุค่าเป็นบวก",
			"first_name":       "โปรดระบุไม่เกิน 50 ตัวอักษร",
			"last_name":        "โปรดระบุไม่เกิน 50 ตัวอักษร",
		},
		"summary debit and credit not equal": {"default": "ผลรวมเดบิตไม่เท่ากับผลรวมเครดิต"},
		"less than max 255 characters":       {"default": "โปรดระบุข้อมูลไม่เกิน 255 ตัวอักษร"},
		"less than min 10 number":            {"default": "โปรดระบุหมายเลขโทรศัพท์ให้ครบ 10 หลัก"},
		"must be integer":                    {"default": "โปรดระบุเป็นตัวเลข"},
		"out of package":                     {"default": "กรุณาเลือก Package ใหม่เนื่องจาก Package ไม่เพียงพอไม่สามารถทำรายการต่อได้"},
		"invalid length":                     {"default": "โปรดระบุตัวเลขให้ครบ 13 หลัก"},
		"invalid Thai National ID":           {"default": "รูปแบบรหัสประชาชนไม่ถูกต้อง"},
	}
}

// รับ parameter 3 ตัว 1.request struct 2.*gin.Context 3.*gorm.DB
//
//tag unique เช็คว่าค่าๆนั้นมีซ้ำหรือไม่ โดยต้องส่งสามค่า `validate:"unique=table_name|column_name|optional_condition"` ถ้าไม่ต้องการเพิ่ม optional_condition ส่งมาแค่สองค่า unique=table_name|column_name
//
// กรณีใช้ tag unique ต้องส่ง *gorm.DB มาด้วย ถ้าไม่มี tag unique ให้ส่ง nil
//
// Ex. กรณีไม่มี tag unique err := helpers.Validate(request,c,nil)
//
// Ex. กรณีมี tag unique err := helpers.Validate(request,c,databases.NewPostgres())
//
//tag acceptlist คือการกำหนดว่าfield นี้จะสามารถใส่ค่าอะไรได้บ้าง Ex. `validate:"acceptlist=asc|dec"`
//
//tag date คือการเช็ค check format yyyy-mm-dd optional `validate:"date=lt"` lt = <= date now
//
//tag validateTel เช็ครูปแบบเบอร์โทร
//
//tag thaiid เช็คเลขบัตรปชช ของคนไทย
//
//tag email เช็ครูปแบบอีเมล์
func Validate(v interface{}, c *gin.Context, db *gorm.DB) *fail {

	validate := initValidator(c, db)

	mapError := map[string]string{}

	if errs := validate.Validate(v); errs != nil {
		errorArray := errs.Error() //message จะบอกมี pattern ว่า FieldName: message error,...
		parseErrors(errorArray, mapError)
		errResponse := ValidateResponse(mapError)
		return &errResponse
	}

	return &fail{}
}

func parseErrors(errorArray string, mapError map[string]string) {
	re := regexp.MustCompile(`\b\w+?\[.*?\]?\.\w+?:|\b\w+?:`)
	listDatas := re.FindAllString(errorArray, -1)

	for index, fieldName := range listDatas {
		var listErr []string
		key := parseFieldName(strings.TrimSuffix(fieldName, ":"))
		lastFieldName := getLastFieldName(strings.TrimSuffix(fieldName, ":"))

		indexFirstElement := strings.Index(errorArray, fieldName) + len(fieldName)
		indexNextElement := len(errorArray)
		if index != len(listDatas)-1 {
			indexNextElement = strings.Index(errorArray, listDatas[index+1])
		}

		errMessages := strings.Split(strings.TrimSpace(errorArray[indexFirstElement:indexNextElement]), ",")
		for _, errMsg := range errMessages {
			if errMsg = strings.TrimSpace(errMsg); errMsg != "" {
				listErr = append(listErr, handleErrMesssage(lo.SnakeCase(lastFieldName), errMsg))
			}
		}

		mapError[key] = strings.Join(listErr, "|")
	}
}

func getLastFieldName(fieldName string) string {
	re := regexp.MustCompile(`\w+$`)
	return re.FindString(fieldName)
}

func parseFieldName(fieldName string) string {
	re := regexp.MustCompile(`(\w+|\[\d+\])`)
	parts := re.FindAllString(fieldName, -1)

	for i, part := range parts {
		if strings.HasPrefix(part, "[") && strings.HasSuffix(part, "]") {
			parts[i] = part // Keep array indices as they are
		} else {
			if i == 0 {
				parts[i] = lo.SnakeCase(part)
			} else {
				parts[i] = "._" + lo.SnakeCase(part)
			}
		}
	}

	return strings.Join(parts, "")
}

// HandleErrMesssage maps error messages to human-readable text
func handleErrMesssage(errField, err string) string {
	errorMessages := initErrorMessages()

	//for accept list validate
	data := strings.Split(err, "&v=")
	if len(data) > 1 {
		sub := data[1]
		return "รองรับเฉพาะ" + " " + sub
	}
	///
	if fieldErrors, exists := errorMessages[err]; exists {
		if msg, ok := fieldErrors[errField]; ok {
			return msg
		}
		if msg, ok := fieldErrors["default"]; ok {
			return msg
		}
	}
	return "เกิดข้อผิดพลาด โปรดลองอีกครั้ง"
}

func thaiNationalIDValidator(v interface{}, param string) error {
	id, ok := v.(string)
	if !ok {
		return errors.New("invalid type")
	}

	if len(id) != 13 {
		return errors.New("invalid length")
	}

	sum := 0
	for i := 0; i < 12; i++ {
		num, err := strconv.Atoi(string(id[i]))
		if err != nil {
			return errors.New("invalid Thai National ID")
		}
		sum += num * (13 - i)
	}

	remainder := sum % 11
	checkDigit := (11 - remainder) % 10

	lastDigit, err := strconv.Atoi(string(id[12]))
	if err != nil {
		return errors.New("invalid Thai National ID")
	}

	if checkDigit != lastDigit {
		return errors.New("invalid Thai National ID")
	}

	return nil
}

// Custom
// example `validate:"acceptlist=asc|dec"`
func acceptList(v interface{}, param string) error {
	st := reflect.ValueOf(v)
	if st.String() == "" {
		return nil
	}
	paramContains := strings.ToUpper("|" + param + "|")
	value := "|" + strings.ToUpper(st.String()+"|")
	if exists := strings.Contains(paramContains, value); !exists {
		return errors.New("ONLY_SUPPORT&v=" + param)
	}

	return nil
}

// Custom
// example `validate:"date"` check format yyyy-mm-dd
// optional `validate:"date=lt"` lt = <= date now
func dateFormat(v interface{}, param string) error {
	date, ok := v.(string)
	if !ok {
		return errors.New("VALID_DATE_FORMAT")
	}
	if date == "" {
		return nil
	}

	dateTime, err := time.Parse("2006-01-02", date)
	if err != nil {
		return errors.New("VALID_DATE_FORMAT")
	}

	if strings.TrimSpace(param) == "lt" {
		now := time.Now()
		if dateTime.After(now) {
			return errors.New("VALID_DATE_OVERNOW")
		}
	}

	return nil
}

// validateUniqueValue returns a validation function to check the uniqueness of a value in the database.
// This function is specifically designed for use with POST and PUT HTTP methods.
// Usage: `validate:"unique=table_name|column_name|optional_condition"`
func validateUniqueValue(c *gin.Context, db *gorm.DB) validator.ValidationFunc {
	return func(v interface{}, param string) error {
		value, ok := v.(string)
		if !ok {
			return fmt.Errorf("validateUniqueValue only validates strings")
		}

		// Split the parameters
		params := strings.Split(param, "|")
		if len(params) != 2 && len(params) != 3 {
			return fmt.Errorf("validateUniqueValue requires 2 or 3 parameters, got %d", len(params))
		}

		// Extract table name, column name, and optional condition
		var tableName, columnName, condition string
		tableName, columnName = params[0], params[1]
		if len(params) == 3 {
			condition = params[2]
		}

		var count int64
		if condition != "" {
			// Check uniqueness with condition
			db.Table(tableName).Where(columnName+" = ? AND "+condition, value).Count(&count)
		} else {
			// Check uniqueness without condition
			db.Table(tableName).Where(columnName+" = ?", value).Count(&count)
		}

		// If count > 0, the value is not unique
		if count > 0 {
			// Handle POST method: simply check for existence
			if c.Request.Method == "POST" {
				return errors.New("already used")
			}
			// Handle PUT method: ensure the value is unique excluding the current record
			if c.Request.Method == "PUT" {
				ID := c.Param("id")
				var existingCount int64
				if condition != "" {
					db.Table(tableName).Where(columnName+" = ? AND id = ? AND "+condition, value, ID).Count(&existingCount)
				} else {
					db.Table(tableName).Where(columnName+" = ? AND id = ?", value, ID).Count(&existingCount)
				}
				if existingCount == 0 {
					return errors.New("already used")
				}
			}
		}

		return nil
	}
}

func validateEmail(v interface{}, param string) error {

	email, ok := v.(string)
	if !ok {
		return validator.ErrUnsupported
	}

	if len(email) == 0 {
		return fmt.Errorf("zero value")
	}
	if validateErr := emailValidate.New().Var(email, "email"); validateErr != nil {
		return errors.New("email: invalid email address, ")
	}
	return nil
}

// Custom validation function for the Tel field
func validateTel(v interface{}, param string) error {
	tel, ok := v.(string)
	if !ok {
		return validator.ErrUnsupported
	}

	if len(tel) == 0 {
		return fmt.Errorf("zero value")
	}

	if len(tel) < 10 { // Assuming 10 is the minimum length
		return fmt.Errorf("less than min")
	}

	// Add more conditions as needed

	return nil
}
