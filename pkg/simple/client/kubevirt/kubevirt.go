/*
Copyright(c) 2024-present Accton. All rights reserved. www.accton.com.tw
*/

package kubevirt

import (
	"github.com/spf13/pflag"
	"kubevirt.io/client-go/kubecli"
)

func NewKubevirtClient(options *Options) (kubecli.KubevirtClient, error) {
	if options.Enable {
		clientConfig := kubecli.DefaultClientConfig(&pflag.FlagSet{})
		return  kubecli.GetKubevirtClientFromClientConfig(clientConfig)
	} else {
		return nil, nil
	}
}
