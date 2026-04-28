package currencies

type CurrencyCode string

type Currency struct {
	ISOCode     int16
	ISOCharCode CurrencyCode
	Name        *string
	LatName     string
	MinorUnits  int32
}

type CurrencyPair struct {
	Base  CurrencyCode
	Quote CurrencyCode
}
