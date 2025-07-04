package model

type Symbol struct {
	ID       int64  `db:"id"`
	Code     string `db:"code"`
	ProdID   int    `db:"prod_id"`
	Base     string `db:"base"`
	Quote    string `db:"quote"`
	Decimals int64  `db:"decimals"`
}
