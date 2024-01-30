package taran

type DCCacheStruct struct {
	Id                  int
	Status              int
	Ttl                 int
	Attempts            int
	Log                 int
	IsoCallEnd          int
	OriginationCarrier  int
	Screenshot          int
	IncomingNumber      int
	DeviceFailed        int
	IncomingCallStatus  int
	FromNum             int
	ToNum               int
	CallEnd             int
	MakeCallStatusCode  int
	DeviceId            int
	MessageId           int
	CallStart           int
	CallDuration        int
	IncomingNumberMatch int
	IsoCallstart        int
	Cnam                int
	Text                int
	TextRecognized      int
	Custom              int
	ConfigId            int
}

//TODO: Add spam pairs fields

func (t *Tarantool) DCCacheFields() *DCCacheStruct {
	return &DCCacheStruct{
		Id:                  0,
		Status:              1,
		Log:                 2,
		IsoCallEnd:          3,
		OriginationCarrier:  4,
		Screenshot:          5,
		IncomingNumber:      6,
		DeviceFailed:        7,
		IncomingCallStatus:  8,
		FromNum:             9,
		ToNum:               10,
		CallEnd:             11,
		MakeCallStatusCode:  12,
		DeviceId:            13,
		MessageId:           14,
		CallStart:           15,
		CallDuration:        16,
		IncomingNumberMatch: 17,
		IsoCallstart:        18,
		Cnam:                19,
		Text:                20,
		TextRecognized:      21,
		Ttl:                 22,
		ConfigId:            23,
	}
}

func (t *Tarantool) DCCache2Fields() *DCCacheStruct {
	return &DCCacheStruct{
		Id:                  0,
		Status:              1,
		Ttl:                 2,
		Attempts:            3,
		Log:                 4,
		IsoCallEnd:          5,
		OriginationCarrier:  6,
		Screenshot:          7,
		IncomingNumber:      8,
		DeviceFailed:        9,
		IncomingCallStatus:  10,
		FromNum:             11,
		ToNum:               12,
		CallEnd:             13,
		MakeCallStatusCode:  14,
		DeviceId:            15,
		MessageId:           16,
		CallStart:           17,
		CallDuration:        18,
		IncomingNumberMatch: 19,
		IsoCallstart:        20,
		Cnam:                21,
		Text:                22,
		TextRecognized:      23,
		Custom:              24,
		ConfigId:            25,
	}
}
