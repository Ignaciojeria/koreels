package entity

type IdempotencyRecord struct {
	IdempotencyKey   string
	TransactionId    string
	RequestHash      string
	ResponsePayload  []byte // JSON cache para reentregar la misma respuesta
}
