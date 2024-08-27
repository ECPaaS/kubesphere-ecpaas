/*
Copyright(c) 2023-present Accton. All rights reserved. www.accton.com.tw
*/

package minio

import (
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"k8s.io/klog"
	volume "kubesphere.io/kubesphere/pkg/kapis/volume/v1alpha1"
)

func NewMinioClient(options *Options) (*volume.MinioClient, error) {

	useSSL := false

	minioClient, err := minio.New(options.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(options.AccessKeyID, options.SecretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		klog.Fatalf("unable to create MinioClient: %v", err)
	}
	client := volume.MinioClient{
		MinioClient: minioClient,
		EndpointURL: options.Endpoint,
		Username:    options.AccessKeyID,
		Password:    options.SecretAccessKey,
	}

	return &client, err
}
