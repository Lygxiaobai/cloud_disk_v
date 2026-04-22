package helper

import (
	"context"
	"crypto/md5"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/smtp"
	"path"
	"strings"
	"time"

	"cloud_disk/core/internal/define"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"github.com/golang-jwt/jwt/v4"
	"github.com/jordan-wright/email"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

const multipartChunkSize = 5 * 1024 * 1024

type MailConfig struct {
	From       string
	Host       string
	Username   string
	Password   string
	ServerName string
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func CheckPassword(hashed, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
}

func GenerateToken(id int, identity, name, role, secret string, expireSeconds int) (string, error) {
	uc := define.UserClaim{
		ID:       id,
		Identity: identity,
		Name:     name,
		Role:     role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Second * time.Duration(expireSeconds)).Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, uc)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func AnalyzeToken(token, secret string) (*define.UserClaim, error) {
	uc := new(define.UserClaim)
	claims, err := jwt.ParseWithClaims(token, uc, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !claims.Valid {
		return nil, errors.New("invalid token")
	}
	return uc, nil
}

func MailCodeSend(userEmail, code string, cfg MailConfig) error {
	e := email.NewEmail()
	e.From = cfg.From
	e.To = []string{userEmail}
	e.Subject = "Verification Code"
	e.HTML = []byte("<h1>" + code + "</h1>")
	return e.SendWithTLS(
		cfg.Host,
		smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.ServerName),
		&tls.Config{InsecureSkipVerify: true, ServerName: cfg.ServerName},
	)
}

func RandCode(length int) string {
	const digits = "1234567890"
	if length <= 0 {
		length = 6
	}
	code := make([]byte, 0, length)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		code = append(code, digits[r.Intn(len(digits))])
	}
	return string(code)
}

func UUID() string {
	return uuid.NewV4().String()
}

func ValidateUploadExt(ext string, blocked []string) error {
	if ext == "" {
		return nil
	}
	normalized := strings.ToLower(ext)
	for _, b := range blocked {
		if strings.ToLower(b) == normalized {
			return fmt.Errorf("upload of %s is not allowed", ext)
		}
	}
	return nil
}

func HashAndReset(file io.ReadSeeker, maxSize int64) (string, error) {
	if file == nil {
		return "", errors.New("file is empty")
	}

	hasher := md5.New()
	reader := io.Reader(file)
	if maxSize > 0 {
		reader = io.LimitReader(file, maxSize+1)
	}
	written, err := io.Copy(hasher, reader)
	if err != nil {
		return "", err
	}
	if maxSize > 0 && written > maxSize {
		return "", errors.New("file exceeds max upload size")
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

func FileUpload(fileName string, reader io.Reader) (string, error) {
	return uploadObject(fileName, reader, "invalid parameters, bucket name required")
}

func FileUploadMultipart(fileName string, reader io.Reader) (string, error) {
	return uploadObject(fileName, reader, "invalid parameters, source bucket name required")
}

func uploadObject(fileName string, reader io.Reader, bucketErr string) (string, error) {
	objectName := "cloud-disk/" + UUID() + path.Ext(fileName)
	if len(define.BucketName) == 0 {
		flag.PrintDefaults()
		return "", errors.New(bucketErr)
	}
	if len(define.Region) == 0 {
		flag.PrintDefaults()
		return "", errors.New("invalid parameters, region required")
	}
	if len(objectName) == 0 {
		flag.PrintDefaults()
		return "", errors.New("invalid parameters, object name required")
	}

	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewEnvironmentVariableCredentialsProvider()).
		WithRegion(define.Region)
	client := oss.NewClient(cfg)

	request := &oss.PutObjectRequest{
		Bucket: oss.Ptr(define.BucketName),
		Key:    oss.Ptr(objectName),
		Body:   reader,
	}
	result, err := client.PutObject(context.TODO(), request)
	if err != nil {
		return "", fmt.Errorf("failed to put object: %w", err)
	}

	log.Printf("put object result:%#v\n", result)
	return "https://" + define.BucketName + ".oss-" + define.Region + ".aliyuncs.com/" + objectName, nil
}

type multipartRange struct {
	PartNumber   int32
	Start        int
	EndExclusive int
}

func buildMultipartRanges(fileSize int, chunkSize int) ([]multipartRange, error) {
	if fileSize <= 0 {
		return nil, errors.New("file buffer is empty")
	}
	if chunkSize <= 0 {
		return nil, errors.New("chunk size must be greater than zero")
	}

	ranges := make([]multipartRange, 0, (fileSize+chunkSize-1)/chunkSize)
	partNumber := int32(1)
	for start := 0; start < fileSize; start += chunkSize {
		end := start + chunkSize
		if end > fileSize {
			end = fileSize
		}
		ranges = append(ranges, multipartRange{
			PartNumber:   partNumber,
			Start:        start,
			EndExclusive: end,
		})
		partNumber++
	}
	return ranges, nil
}
