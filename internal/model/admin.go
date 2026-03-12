package model

type DashboardSummary struct {
	TotalSalesAmount float64 `json:"totalSalesAmount"`
	TotalSalesCount  int     `json:"totalSalesCount"`
	QrTopupAmount    float64 `json:"qrTopupAmount"`
	CashTopupAmount  float64 `json:"cashTopupAmount"`
	CashRefundAmount float64 `json:"cashRefundAmount"`
}

type StallRevenue struct {
	ShopName    string  `db:"stall_name" json:"stall_name"`
	ShopID      string  `json:"shopId"`
	ShopType    string  `json:"shopType"`
	PaymentAmt  float64 `json:"paymentAmt"`
	GPAmount    float64 `json:"gpAmount"`
	VatAmount   float64 `json:"vatAmount"`
	NetIncome   float64 `json:"netIncome"`
	TotalAmount float64 `db:"total_amount" json:"total_amount"`
}
