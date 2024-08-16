/*
Copyright(c) 2024-present Accton. All rights reserved. www.accton.com.tw
*/

package kubevirt

import (
	"github.com/spf13/pflag"

	"kubesphere.io/kubesphere/pkg/utils/reflectutils"
)

type Options struct {
	Enable bool `json:"enable" yaml:"enable"`
}

func NewKubevirtOptions() *Options {
	return &Options{
		Enable: true,
	}
}

func (s *Options) ApplyTo(options *Options) {
	if s.Enable != options.Enable {
		reflectutils.Override(options, s)
	}
}

func (s *Options) Validate() []error {
	errs := make([]error, 0)
	return errs
}

func (s *Options) AddFlags(fs *pflag.FlagSet, c *Options) {
	fs.BoolVar(&s.Enable, "kubevirt-enabled", c.Enable, "Enable kubevirt component or not.")
}
