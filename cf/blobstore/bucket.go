package blobstore

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
)

func Upload(blob io.Reader, key string, contentType string) (r *s3manager.UploadOutput, err error) {
	uploader := s3manager.NewUploader(Session)
	svc := uploader.S3.(*s3.S3)
	svc.Handlers.Sign.Clear()
	svc.Handlers.Sign.PushBack(SignV2)

	r, err = uploader.Upload(&s3manager.UploadInput{
		Body:        blob,
		Bucket:      &BucketName,
		Key:         &key,
		ContentType: &contentType,
	})

	return
}

func Put(blob io.ReadSeeker, key string, contentType string) (r *s3.PutObjectOutput, err error) {
	r, err = S3.PutObject(&s3.PutObjectInput{
		Body:        blob,
		Bucket:      &BucketName,
		Key:         &key,
		ContentType: &contentType,
	})
	return
}

func List() (files []string) {
	params := &s3.ListObjectsInput{
		Bucket: &BucketName,
	}

	r, err := S3.ListObjects(params)
	if err != nil {
		return
	}

	for _, file := range r.Contents {
		files = append(files, *file.Key)
	}

	return
}

func Get(key string) (r *s3.GetObjectOutput, err error) {
	input := &s3.GetObjectInput{
		Bucket: &BucketName,
		Key:    &key,
	}

	r, err = S3.GetObject(input)
	return
}

func Delete(key string) (r *s3.DeleteObjectOutput, err error) {
	params := &s3.DeleteObjectInput{
		Bucket: &BucketName,
		Key:    &key,
	}

	r, err = S3.DeleteObject(params)
	return
}
