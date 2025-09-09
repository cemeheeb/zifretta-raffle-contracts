package tracker

const MarketplaceAddressRaw = "0:584ee61b2dff0837116d0fcb5078d93964bcbe9c05fd6a141b1bfca5d6a43e18"
const GlobalDeployedTimeout = 300
const GlobalLimitWindowSize = 50

var WalletMap = map[string]int{
	"V1R1":         0,
	"V1R2":         1,
	"V1R3":         2,
	"V2R1":         3,
	"V2R2":         4,
	"V3R1":         5,
	"V3R2":         6,
	"V3R2Lockup":   7,
	"V4R1":         8,
	"V4R2":         9,
	"V5Beta":       10,
	"V5R1":         11,
	"HighLoadV1R1": 12,
	"HighLoadV1R2": 13,
	"HighLoadV2":   14,
	"HighLoadV2R1": 15,
	"HighLoadV2R2": 16,
}
