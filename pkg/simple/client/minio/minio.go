/*
Copyright(c) 2023-present Accton. All rights reserved. www.accton.com.tw
*/

package minio

import (
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"k8s.io/klog"
)

func NewMinioClient(options *Options) (*minio.Client, error) {

	useSSL := false

	minioClient, err := minio.New(options.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(options.AccessKeyID, options.SecretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		klog.Fatalf("unable to create MinioClient: %v", err)
	}

	return minioClient, err
}
