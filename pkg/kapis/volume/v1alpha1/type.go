/*
Copyright(c) 2024-present Accton. All rights reserved. www.accton.com.tw
*/

package v1

import "github.com/minio/minio-go/v7"

type MinioClient struct {
	MinioClient *minio.Client
	EndpointURL string
	Username    string
	Password    string
}
