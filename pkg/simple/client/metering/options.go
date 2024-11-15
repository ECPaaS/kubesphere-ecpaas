package metering

type PriceInfo struct {
	// currency unit, currently support CNY and USD
	CpuPerCorePerHour float64 `json:"cpuPerCorePerHour" yaml:"cpuPerCorePerHour"`
	// cpu cost with above currency unit for per core per hour
	MemPerGigabytesPerHour float64 `json:"memPerGigabytesPerHour" yaml:"memPerGigabytesPerHour"`
	// mem cost with above currency unit for per GB per hour
	IngressNetworkTrafficPerMegabytesPerHour float64 `json:"ingressNetworkTrafficPerMegabytesPerHour" yaml:"ingressNetworkTrafficPerGiagabytesPerHour"`
	// ingress network traffic cost with above currency unit for per MB per hour
	EgressNetworkTrafficPerMegabytesPerHour float64 `json:"egressNetworkTrafficPerMegabytesPerHour" yaml:"egressNetworkTrafficPerGigabytesPerHour"`
	// egress network traffice cost with above currency unit for per MB per hour
	PvcPerGigabytesPerHour float64 `json:"pvcPerGigabytesPerHour" yaml:"pvcPerGigabytesPerHour"`
	// pvc cost with above currency unit for per GB per hour
	CurrencyUnit string `json:"currencyUnit" yaml:"currencyUnit"`
	// gpu cost with above currency unit for per percentage per hour
	GpuPerPercentagePerHour float64 `json:"gpuPerPercentagePerHour" yaml:"gpuPerPercentagePerHour"`
	// gpu framebuffer cost with above currency unit for per MB per hour
	GpuFbPerMegabytesPerHour float64 `json:"gpuFbPerMegabytesPerHour" yaml:"gpuFbPerMegabytesPerHour"`
	// gpu power cost with above currency unit for per watt per hour
	GpuPowerPerWattPerHour float64 `json:"gpuPowerPerWattPerHour" yaml:"gpuPowerPerWattPerHour"`
	// gpu memory cost with above currency unit for per percentage per hour
	GpuMemPerPercentagePerHour float64 `json:"gpuMemPerPercentagePerHour" yaml:"gpuMemPerPercentagePerHour"`
}

type Billing struct {
	PriceInfo PriceInfo `json:"priceInfo" yaml:"priceInfo"`
}

type Options struct {
	RetentionDay string  `json:"retentionDay" yaml:"retentionDay"`
	Billing      Billing `json:"billing" yaml:"billing"`
}

var DefaultMeteringOption = Options{
	RetentionDay: "7d",
	Billing: Billing{
		PriceInfo: PriceInfo{
			CpuPerCorePerHour:                        0,
			MemPerGigabytesPerHour:                   0,
			IngressNetworkTrafficPerMegabytesPerHour: 0,
			EgressNetworkTrafficPerMegabytesPerHour:  0,
			PvcPerGigabytesPerHour:                   0,
			CurrencyUnit:                             "",
			GpuPerPercentagePerHour:                  0,
			GpuFbPerMegabytesPerHour:                 0,
			GpuPowerPerWattPerHour:                   0,
			GpuMemPerPercentagePerHour:               0,
		},
	},
}

func NewMeteringOptions() *Options {
	return &Options{}
}
