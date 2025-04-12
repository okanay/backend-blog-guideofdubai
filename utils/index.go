package utils

import (
	"database/sql"
	"log"
	"math/rand"
	"net/http"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

const TokenAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

func GenerateRandomString(length int) string {
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = TokenAlphabet[seededRand.Intn(len(TokenAlphabet))]
	}

	return string(b)
}

func GenerateRandomInt(min, max int) int {
	return rand.Intn(max-min) + min
}

// EncryptPassword şifreyi geri döndürülemeyecek şekilde şifreler.
func EncryptPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword password ve hash karşılaştırır.
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// TimeTrack zamanı takip eder.
func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)

	if elapsed <= 5*time.Millisecond {
		return
	}

	log.Printf("%s ~TOOK~ %s", name, elapsed.Round(time.Millisecond))
}

// ScanStructByDBTags veritabanı etiketlerine göre structları tarar.
func ScanStructByDBTags(rows *sql.Row, dest interface{}) error {
	v := reflect.ValueOf(dest).Elem()
	fields := make([]interface{}, v.NumField())

	for i := 0; i < v.NumField(); i++ {
		tag := v.Type().Field(i).Tag.Get("db")
		if tag != "" && tag != "-" {
			fields[i] = v.Field(i).Addr().Interface()
		}
	}

	return rows.Scan(fields...)
}

// ScanStructByDBTagsForRows veritabanı etiketlerine göre structları tarar.
func ScanStructByDBTagsForRows(rows *sql.Rows, dest interface{}) error {
	v := reflect.ValueOf(dest).Elem()
	fields := make([]interface{}, v.NumField())

	for i := 0; i < v.NumField(); i++ {
		tag := v.Type().Field(i).Tag.Get("db")
		if tag != "" && tag != "-" {
			fields[i] = v.Field(i).Addr().Interface()
		}
	}

	return rows.Scan(fields...)
}

// ValidateRequest veritabanı etiketlerine göre structları tarar.
func ValidateRequest(c *gin.Context, req interface{}) error {
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request.", "message": err.Error()})
		return err
	}

	validate := validator.New()
	err := validate.Struct(req)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}

	return nil
}

// ValidateRequestWithGinRequest veritabanı etiketlerine göre structları tarar.
func ValidateRequestWithGinRequest(c *gin.Context, req interface{}) error {
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request.", "message": err.Error()})
		return err
	}

	validate := validator.New()
	err := validate.Struct(req)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}

	return nil
}
