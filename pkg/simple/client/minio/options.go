/*
Copyright 2020 KubeSphere Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package minio

import (
	"github.com/spf13/pflag"
	"kubesphere.io/kubesphere/pkg/utils/reflectutils"
)

// Options contains configuration to access a s3 service
type Options struct {
	Endpoint        string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	AccessKeyID     string `json:"accessKeyID,omitempty" yaml:"accessKeyID,omitempty"`
	SecretAccessKey string `json:"secretAccessKey,omitempty" yaml:"secretAccessKey,omitempty"`
	Bucket          string `json:"bucket,omitempty" yaml:"bucket,omitempty"`
}

// NewS3Options creates a default disabled Options(empty endpoint)
func NewMinioOptions() *Options {
	return &Options{
		Endpoint:        "minio.kubesphere-system.svc:9000",
		AccessKeyID:     "openpitrixminioaccesskey",
		SecretAccessKey: "openpitrixminiosecretkey",
		Bucket:          "ecpaas-images",
	}
}

// Validate check options values
func (s *Options) Validate() []error {
	var errors []error

	return errors
}

// ApplyTo overrides options if it's valid, which endpoint is not empty
func (s *Options) ApplyTo(options *Options) {
	if s.Endpoint != "" {
		reflectutils.Override(options, s)
	}
}

// AddFlags add options flags to command line flags,
// if minio-endpoint if left empty, following options will be ignored
func (s *Options) AddFlags(fs *pflag.FlagSet, c *Options) {
	fs.StringVar(&s.Endpoint, "minio-endpoint", c.Endpoint, ""+
		"Endpoint to access to minio object storage service, if left blank, the following options "+
		"will be ignored.")

	fs.StringVar(&s.AccessKeyID, "minio-access-key-id", c.AccessKeyID, "access key of minio")

	fs.StringVar(&s.SecretAccessKey, "minio-secret-access-key", c.SecretAccessKey, "secret access key of minio")

	fs.StringVar(&s.Bucket, "minio-bucket", c.Bucket, "bucket name of minio")
}
