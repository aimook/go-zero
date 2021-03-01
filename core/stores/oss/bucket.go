package oss

import (
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/tal-tech/go-zero/core/logx"
	"mime/multipart"
)

type ObjectFile struct {
	Bucket   string
	Name     string
	Size     int64
	Content  multipart.File
	MimeType string
}

type ObjectInfo struct {
	minio.UploadInfo
}

func (oc *OSSClient) checkBucketIsExists(bucket string) bool {
	isExists, error := oc.client.BucketExists(context.Background(), bucket)
	if error == nil && isExists {
		logx.Info("%s already exists", bucket)
		return true
	} else {
		logx.Error(error)
		return false
	}
}

func (oc *OSSClient) CreateBucket(bucket string) {
	isExists := oc.checkBucketIsExists(bucket)
	if isExists {
		logx.Infof("bucket %s is existed", bucket)
		return
	}
	err := oc.client.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{
		Region:        "cn-east-1",
		ObjectLocking: false,
	})
	if err != nil {
		logx.Error(err)
	}
}

func (oc *OSSClient) PutObject(object ObjectFile) (ObjectInfo, error) {
	uploadInfo, err := oc.client.PutObject(context.Background(), object.Bucket, object.Name, object.Content, object.Size, minio.PutObjectOptions{
		ContentType: object.MimeType,
	})
	if err != nil {
		logx.Error(err)
		return ObjectInfo{}, err
	}
	logx.Debugf("%v", uploadInfo)
	return ObjectInfo{uploadInfo}, nil
}

func (oc *OSSClient) ListBuckets() ([]string, error) {
	buckets, err := oc.client.ListBuckets(context.Background())
	if err != nil {
		logx.Error(err)
		return []string{}, err
	}
	bucketNames := make([]string, len(buckets))
	for _, bucket := range buckets {
		bucketNames = append(bucketNames, bucket.Name)
		logx.Debug(bucket)
	}
	return bucketNames, nil
}
