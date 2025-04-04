package acs3

import (
	"log"
	"mime/multipart"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/google/uuid"
)

var sess *session.Session
var svc *s3.S3

// initiate a connection to s3
func InitConnection(AccessKey, SecretKey, S3_Region, S3_Endpoint string) {
	var err error

	sess, err = session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(AccessKey, SecretKey, ""),
	})

	svc = s3.New(sess, &aws.Config{
		Region:   aws.String(S3_Region),
		Endpoint: aws.String(S3_Endpoint),
	})

	if err != nil {
		log.Panicln(err.Error())
	} else {
		log.Println("Successfully created session.")
	}
}

// returns list of existing buckets
func ListBuckets() ([]*s3.Bucket, error) {
	result, err := svc.ListBuckets(nil)
	if err != nil {
		return nil, err
	}

	return result.Buckets, nil
}

// returns list of existing objects
func ListObjects(bucketName string) ([]*s3.Object, error) {
	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return nil, err
	}

	return resp.Contents, nil
}

// download an object from storage
func GetObject(bucketName, objectKey string) ([]byte, error) {
	downloader := s3manager.NewDownloaderWithClient(svc)

	buffer := aws.NewWriteAtBuffer([]byte{})

	_, err := downloader.Download(buffer,
		&s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(objectKey),
		})

	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), err
}

// uploads file end returns generated file key.
func UploadObject(bucketName string, fileHeader *multipart.FileHeader) (string, error) {
	uploader := s3manager.NewUploaderWithClient(svc)

	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	rand, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	hash := rand.String()

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(hash),
		Body:   file,
	})

	if err != nil {
		return "", err
	}

	log.Println("file uploaded: ", hash)
	return hash, err
}

// removes an object with given key
func DeleteObject(bucketName, objectKey string) error {
	_, err := svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})

	return err
}
