package help2pay

// currency => bank code
var depositBanks = map[string]map[string]struct{}{
	"MYR": {
		"AFF":          {},
		"ALB":          {},
		"AMB":          {},
		"BIMB":         {},
		"BSN":          {},
		"CIMB":         {},
		"HLB":          {},
		"HSBC":         {},
		"MBB":          {},
		"OCBC":         {},
		"PBB":          {},
		"RHB":          {},
		"UOB":          {},
		"DUITNOW":      {},
		"TNGODUITNOW":  {},
		"GRABDUITNOW":  {},
		"MAEDUITNOW":   {},
		"BOOSTDUITNOW": {},
	},

	"THB": {
		"BBL":    {},
		"BOA":    {},
		"KKR":    {},
		"KNK":    {},
		"KTB":    {},
		"SCB":    {},
		"TMB":    {},
		"PPTP":   {},
		"TSTB":   {},
		"BBLLBT": {},
		"BOALBT": {},
		"KKRLBT": {},
		"SCBLBT": {},
		"TMBLBT": {},
	},

	"VND": {
		"ACB":           {},
		"AGB":           {},
		"BIDV":          {},
		"DAB":           {},
		"EXIM":          {},
		"HDB":           {},
		"MB":            {},
		"MTMB":          {},
		"OCB":           {},
		"SACOM":         {},
		"TCB":           {},
		"TPB":           {},
		"VCB":           {},
		"VIB":           {},
		"VPB":           {},
		"VTB":           {},
		"VIETQR":        {},
		"VIETQRMOMO":    {},
		"VIETQRZALO":    {},
		"VIETQRVIETTEL": {},
		"VSTB":          {},
		"VCBLBT":        {},
		"DABLBT":        {},
	},

	"PHP": {
		"BDO":       {},
		"BPI":       {},
		"LBP":       {},
		"RCBC":      {},
		"SBC":       {},
		"QRPH":      {},
		"BPIQRPH":   {},
		"EWBQRPH":   {},
		"RCBCQRPH":  {},
		"UBPQRPH":   {},
		"MTBQRPH":   {},
		"GCASHQRPH": {},
		"PSTB":      {},
		"BPILBT":    {},
	},

	"INR": {
		"AXIS":     {},
		"HDFC":     {},
		"IDFC":     {},
		"INDUSIND": {},
		"KOTAK":    {},
		"YES":      {},
		"CIUB":     {},
		"FEDERAL":  {},
		"IDIB":     {},
		"ICICI":    {},
		"UPI":      {},
		"HDFCUPI":  {},
	},

	"IDR": {
		"BCA":         {},
		"BNI":         {},
		"BRI":         {},
		"CIMBN":       {},
		"MDR":         {},
		"PMTB":        {},
		"PANIN":       {},
		"QRIS":        {},
		"DANAQRIS":    {},
		"GOPAYQRIS":   {},
		"LINKAJAQRIS": {},
		"OVOQRIS":     {},
		"SHOPEEQRIS":  {},
		"ISTB":        {},
		"BCAVA":       {},
		"BNIVA":       {},
		"BRIVA":       {},
		"CIMBNVA":     {},
		"MBBIVA":      {},
		"MDRVA":       {},
		"PMTBVA":      {},
		"PANINVA":     {},
		"IMTB":        {},
		"BCALBT":      {},
		"PMTBLBT":     {},
	},
}

// bank code ==> bank name
var bankCodes = map[string]string{
	"AFF":           "Affin Bank",
	"ALB":           "Alliance Bank Malaysia Berhad",
	"AMB":           "AmBank Group",
	"BIMB":          "Bank Islam Malaysia Berhad",
	"BSN":           "Bank Simpanan Nasional",
	"CIMB":          "CIMB Bank Berhad",
	"HLB":           "Hong Leong Bank Berhad",
	"HSBC":          "HSBC Bank (Malaysia) Berhad",
	"MBB":           "Maybank Berhad",
	"OCBC":          "OCBC Bank (Malaysia) Berhad",
	"PBB":           "Public Bank Berhad",
	"RHB":           "RHB Banking Group",
	"UOB":           "United Overseas Bank (Malaysia) Bhd",
	"DUITNOW":       "Duitnow",
	"TNGODUITNOW":   "Touch N Go",
	"GRABDUITNOW":   "GrabPay",
	"MAEDUITNOW":    "MAE",
	"BOOSTDUITNOW":  "BOOST",
	"BBL":           "Bangkok Bank",
	"BOA":           "Bank of Ayudhya (Krungsri)",
	"KKR":           "Karsikorn Bank (K-Bank)",
	"KNK":           "Kiatnakin Bank",
	"KTB":           "Krung Thai Bank",
	"SCB":           "Siam Commercial Bank",
	"TMB":           "TMBThanachart Bank(TTB)",
	"PPTP":          "Promptpay",
	"TSTB":          "Thai Semi Transfer Bank",
	"BBLLBT":        "BBL Local Bank Transfer",
	"BOALBT":        "BOA Local Bank Transfer",
	"KKRLBT":        "KKR Local Bank Transfer",
	"SCBLBT":        "SCB Local Bank Transfer",
	"TMBLBT":        "TMB Local Bank Transfer",
	"ACB":           "Asia Commercial Bank",
	"AGB":           "Agribank",
	"BIDV":          "Bank for Investment and Development of Vietnam",
	"DAB":           "DongA Bank",
	"EXIM":          "Eximbank Vietnam",
	"HDB":           "HDB Bank",
	"MB":            "Military Commercial Joint Stock Bank",
	"MTMB":          "Maritime Bank",
	"OCB":           "Orient Commercial Joint Stock Bank",
	"SACOM":         "Sacombank",
	"TCB":           "Techcombank",
	"TPB":           "Tien Phong Bank",
	"VCB":           "Vietcombank",
	"VIB":           "Vietnam International Bank",
	"VPB":           "VP Bank",
	"VTB":           "Vietinbank",
	"VIETQR":        "VietQRpay",
	"VIETQRMOMO":    "VietQR MOMO",
	"VIETQRZALO":    "VietQR Zalo Pay",
	"VIETQRVIETTEL": "VietQR Viettel Pay",
	"VSTB":          "VND Semi Transfer Bank",
	"VCBLBT":        "Vietcom Bank Local Bank Transfer",
	"DABLBT":        "Donga Bank Local Bank Transfer",
	"BDO":           "Banco de Oro",
	"BPI":           "Bank of the Philippine Islands",
	"LBP":           "Land Bank of the Philippines",
	"RCBC":          "Rizal Commercial Banking Corporation",
	"SBC":           "Security Bank Corporation",
	"QRPH":          "QRPH",
	"BPIQRPH":       "Bank of the Philippine Islands QRPH",
	"EWBQRPH":       "Eastwest bank QRPH",
	"RCBCQRPH":      "Rizal Commercial Banking Corporation QRPH",
	"UBPQRPH":       "Union Bank of the Philippines QRPH",
	"MTBQRPH":       "Metropolitan Bank & Trust Company QRPH",
	"GCASHQRPH":     "GCASH QRPH",
	"PSTB":          "PHP Semi Transfer Bank",
	"BPILBT":        "Bank of the Philippine Islands LBT",
	"AXIS":          "AXIS Bank",
	"HDFC":          "HDFC Bank",
	"IDFC":          "IDFC Bank",
	"INDUSIND":      "INDUSIND Bank",
	"KOTAK":         "KOTAK Mahindra Bank",
	"YES":           "YES Bank",
	"CIUB":          "CITY UNION",
	"FEDERAL":       "Federal Bank",
	"IDIB":          "INDIAN BANK",
	"ICICI":         "ICICI Bank Limited",
	"UPI":           "Unified Payments Interface",
	"HDFCUPI":       "HDFC Bank UPI",
	"BCA":           "Bank Central Asia",
	"BNI":           "Bank Negara Indonesia",
	"BRI":           "Bank Rakyat Indonesia",
	"CIMBN":         "CIMB Niaga",
	"MDR":           "Mandiri Bank",
	"PMTB":          "Permata Bank",
	"PANIN":         "Panin Bank",
	"QRIS":          "QRIS",
	"DANAQRIS":      "DANA QRIS",
	"GOPAYQRIS":     "GO PAY QRIS",
	"LINKAJAQRIS":   "LINK AJA QRIS",
	"OVOQRIS":       "OVO QRIS",
	"SHOPEEQRIS":    "Shopee Pay QRIS",
	"ISTB":          "IDR Virtual Account",
	"BCAVA":         "BCA Virtual Account",
	"BNIVA":         "BNI Virtual Account",
	"BRIVA":         "BRI Virtual Account",
	"CIMBNVA":       "CIMBN Virtual Account",
	"MBBIVA":        "MBBI Virtual Account",
	"MDRVA":         "MDR Virtual Account",
	"PMTBVA":        "PMTB Virtual Account",
	"PANINVA":       "PANIN Virtual Account",
	"IMTB":          "IDR Manual Transfer Bank",
	"BCALBT":        "BCA Bank Local Bank Transfer",
	"PMTBLBT":       "Bank Permata Local Bank Transfer",
}

type Bank struct {
	Code string
	Name string
}

var fiat_support_banks map[string]*Bank
var currency_support_banks map[CurrencyCode][]string

func init() {
	for code, name := range bankCodes {
		item := &Bank{
			Code: code,
			Name: name,
		}
		fiat_support_banks[code] = item
	}
	for currency, banks := range depositBanks {
		currency_support_banks[currency] = make([]string, 0, len(banks))
		for code := range banks {
			currency_support_banks[currency] = append(currency_support_banks[currency], code)
		}
	}
}

func GetFiatSupportBanks() map[string]*Bank {
	return fiat_support_banks
}

func GetCurrencySupportBanks() map[CurrencyCode][]string {
	return currency_support_banks
}

func IsFiatSupportBank(code string) bool {
	_, ok := fiat_support_banks[code]
	return ok
}

func IsCurrencySupportBank(currency CurrencyCode, code string) bool {
	banks, ok := depositBanks[currency]
	if !ok {
		return false
	}
	_, ok = banks[code]
	return ok
}
