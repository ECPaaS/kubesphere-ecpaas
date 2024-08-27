/*
Copyright(c) 2024-present Accton. All rights reserved. www.accton.com.tw
*/

package kubevirt

import (
	"github.com/spf13/pflag"
	"kubevirt.io/client-go/kubecli"
)

func NewKubevirtClient() (kubecli.KubevirtClient, error) {
	clientConfig := kubecli.DefaultClientConfig(&pflag.FlagSet{})
	return  kubecli.GetKubevirtClientFromClientConfig(clientConfig)
}
