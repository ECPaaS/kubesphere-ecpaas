/*
Copyright(c) 2024-present Accton. All rights reserved. www.accton.com.tw
*/

package v1alpha4

import (
	"github.com/emicklei/go-restful"

	"kubesphere.io/kubesphere/pkg/models/metering"
)

func (h *tenantHandler) HandlePriceInfoQuery(req *restful.Request, resp *restful.Response) {

	var priceResponse metering.PriceResponse

	priceInfo := h.meteringOptions.Billing.PriceInfo
	priceResponse.RetentionDay = h.meteringOptions.RetentionDay
	priceResponse.Currency = priceInfo.CurrencyUnit
	priceResponse.CpuPerCorePerHour = priceInfo.CpuPerCorePerHour
	priceResponse.MemPerGigabytesPerHour = priceInfo.MemPerGigabytesPerHour
	priceResponse.IngressNetworkTrafficPerMegabytesPerHour = priceInfo.IngressNetworkTrafficPerMegabytesPerHour
	priceResponse.EgressNetworkTrafficPerMegabytesPerHour = priceInfo.EgressNetworkTrafficPerMegabytesPerHour
	priceResponse.PvcPerGigabytesPerHour = priceInfo.PvcPerGigabytesPerHour
	priceResponse.GpuPerPercentagePerHour = priceInfo.GpuPerPercentagePerHour
	priceResponse.GpuFbPerMegabytesPerHour = priceInfo.GpuFbPerMegabytesPerHour
	priceResponse.GpuPowerPerKilowattPerHour = priceInfo.GpuPowerPerKilowattPerHour
	priceResponse.GpuMemPerPercentagePerHour = priceInfo.GpuMemPerPercentagePerHour

	resp.WriteAsJson(priceResponse)

	return
}
