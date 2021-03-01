package oss

import (
	"errors"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"log"
)

var (
	ErrInitClientErr = errors.New("init oss client error")

	defaultMinioConf = OSSConf{
		Token:     "",
		EnableSSL: false,
		Region:    "",
	}
)

type (
	OSSClient struct {
		client *minio.Client
	}
	OSSConf struct {
		Endpoint        string
		AccessKeyID     string
		SecretAccessKey string
		Token           string `json:",optional"`
		EnableSSL       bool   `json:",default=false,options=true|false"`
		Region          string `json:",optional"`
	}
)

func NewClientWith(conf OSSConf) *OSSClient {
	//if conf == nil {
	//logx.Info("oss config file is null")
	//conf = defaultMinioConf
	//conf.EnableSSL = defaultMinioConf.EnableSSL
	//conf.Region = defaultMinioConf.Region
	//conf.Token = defaultMinioConf.Token
	//}
	client, err := minio.New(conf.Endpoint, &minio.Options{
		Creds:        credentials.NewStaticV4(conf.AccessKeyID, conf.SecretAccessKey, conf.Token),
		Secure:       conf.EnableSSL,
		Transport:    nil,
		Region:       conf.Region,
		BucketLookup: 0,
		CustomMD5:    nil,
		CustomSHA256: nil,
	})
	if err != nil {
		log.Fatalln(err)
	}
	return &OSSClient{
		client: client,
	}
}

func newClient(conf OSSConf) *minio.Client {
	client, err := minio.New(conf.Endpoint, &minio.Options{
		Creds:        credentials.NewStaticV4(conf.AccessKeyID, conf.SecretAccessKey, conf.Token),
		Secure:       conf.EnableSSL,
		Transport:    nil,
		Region:       conf.Region,
		BucketLookup: 0,
		CustomMD5:    nil,
		CustomSHA256: nil,
	})
	if err != nil {
		log.Fatalln(err)
	}
	return client
}

func (mc OSSConf) NewClient() *minio.Client {
	return newClient(mc)
}
