package helper

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/smtp"
	"path"
	"sort"
	"sync"
	"time"

	"cloud_disk/core/internal/define"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"github.com/golang-jwt/jwt/v4"
	"github.com/jordan-wright/email"
	uuid "github.com/satori/go.uuid"
)

const multipartChunkSize = 5 * 1024 * 1024

func MD5(str string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(str)))
}

func GenerateToken(id int, identity string, name string, role string, expireTime int) (string, error) {
	uc := define.UserClaim{
		ID:       id,
		Identity: identity,
		Name:     name,
		Role:     role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Second * time.Duration(expireTime)).Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, uc)
	tokenString, err := token.SignedString([]byte(define.JwtKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func MailCodeSend(userEmail string, code string) error {
	e := email.NewEmail()
	e.From = "Jordan Wright <18163688304@163.com>"
	e.To = []string{userEmail}
	e.Subject = "Verification Code"
	e.HTML = []byte("<h1>" + code + "</h1>")
	return e.SendWithTLS(
		"smtp.163.com:465",
		smtp.PlainAuth("", "18163688304@163.com", define.MailPassword, "smtp.163.com"),
		&tls.Config{InsecureSkipVerify: true, ServerName: "smtp.163.com"},
	)
}

func RandCode() string {
	const digits = "1234567890"

	code := make([]byte, 0, define.CodeLen)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < define.CodeLen; i++ {
		code = append(code, digits[r.Intn(len(digits))])
	}
	return string(code)
}

func UUID() string {
	return uuid.NewV4().String()
}

func FileUpload(r *http.Request) (string, error) {
	file, header, err := r.FormFile("file")
	if err != nil {
		return "", err
	}
	defer file.Close()

	objectName := "cloud-disk/" + UUID() + path.Ext(header.Filename)
	if len(define.BucketName) == 0 {
		flag.PrintDefaults()
		return "", errors.New("invalid parameters, bucket name required")
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
		Body:   file,
	}
	result, err := client.PutObject(context.TODO(), request)
	if err != nil {
		return "", fmt.Errorf("failed to put object: %w", err)
	}

	log.Printf("put object result:%#v\n", result)
	return "https://" + define.BucketName + ".oss-" + define.Region + ".aliyuncs.com/" + objectName, nil
}

func AnalyzeToken(token string) (*define.UserClaim, error) {
	uc := new(define.UserClaim)
	claims, err := jwt.ParseWithClaims(token, uc, func(token *jwt.Token) (interface{}, error) {
		return []byte(define.JwtKey), nil
	})
	if err != nil {
		return nil, err
	}
	if !claims.Valid {
		return nil, errors.New("invalid token")
	}
	return uc, nil
}

// 文件分片上传
func FileUploadMultipart(fileName string, fileBuf []byte) (string, error) {
	if len(fileBuf) == 0 {
		return "", errors.New("file buffer is empty")
	}

	objectName := "cloud-disk/" + UUID() + path.Ext(fileName)
	if len(define.BucketName) == 0 {
		flag.PrintDefaults()
		return "", errors.New("invalid parameters, source bucket name required")
	}
	if len(define.Region) == 0 {
		flag.PrintDefaults()
		return "", errors.New("invalid parameters, region required")
	}
	if len(objectName) == 0 {
		flag.PrintDefaults()
		return "", errors.New("invalid parameters, source object name required")
	}

	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewEnvironmentVariableCredentialsProvider()).
		WithRegion(define.Region)
	client := oss.NewClient(cfg)

	initRequest := &oss.InitiateMultipartUploadRequest{
		Bucket: oss.Ptr(define.BucketName),
		Key:    oss.Ptr(objectName),
	}
	initResult, err := client.InitiateMultipartUpload(context.TODO(), initRequest)
	if err != nil {
		return "", fmt.Errorf("failed to initiate multipart upload: %w", err)
	}
	if initResult.UploadId == nil {
		return "", errors.New("multipart upload id is empty")
	}

	log.Printf("initiate multipart upload result:%#v\n", *initResult.UploadId)
	uploadID := *initResult.UploadId

	chunks, err := buildMultipartRanges(len(fileBuf), multipartChunkSize)
	if err != nil {
		return "", err
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	parts := make([]oss.UploadPart, 0, len(chunks))
	errCh := make(chan error, len(chunks))

	for _, chunk := range chunks {
		wg.Add(1)
		go func(chunk multipartRange) {
			defer wg.Done()

			partRequest := &oss.UploadPartRequest{
				Bucket:     oss.Ptr(define.BucketName),
				Key:        oss.Ptr(objectName),
				PartNumber: chunk.PartNumber,
				UploadId:   oss.Ptr(uploadID),
				Body:       bytes.NewReader(fileBuf[chunk.Start:chunk.EndExclusive]),
			}

			partResult, err := client.UploadPart(context.TODO(), partRequest)
			if err != nil {
				errCh <- fmt.Errorf("failed to upload part %d: %w", chunk.PartNumber, err)
				return
			}

			mu.Lock()
			parts = append(parts, oss.UploadPart{
				PartNumber: partRequest.PartNumber,
				ETag:       partResult.ETag,
			})
			mu.Unlock()
		}(chunk)
	}

	wg.Wait()
	close(errCh)
	for uploadErr := range errCh {
		if uploadErr != nil {
			return "", uploadErr
		}
	}

	sort.Slice(parts, func(i, j int) bool {
		return parts[i].PartNumber < parts[j].PartNumber
	})

	request := &oss.CompleteMultipartUploadRequest{
		Bucket:   oss.Ptr(define.BucketName),
		Key:      oss.Ptr(objectName),
		UploadId: oss.Ptr(uploadID),
		CompleteMultipartUpload: &oss.CompleteMultipartUpload{
			Parts: parts,
		},
	}
	_, err = client.CompleteMultipartUpload(context.TODO(), request)
	if err != nil {
		return "", fmt.Errorf("failed to complete multipart upload: %w", err)
	}

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
