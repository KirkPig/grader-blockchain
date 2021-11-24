package services

type AuthorizationRequest struct {
	PublicKey string `json:"publicKey" binding:"required"`
	SecretKey string `json:"secretKey" binding:"required"`
	StudentId string `json:"studentId" binding:"required"`
	Pin       string `json:"pin" binding:"required"`
}

type AuthorizationResponse struct {
	Status          string `json:"status" binding:"required"`
	TransactionHash string `json:"transactionHash" binding:"required"`
	ErrorLog        string `json:"errorLog",omitempty binding:"required"`
}

type Transaction struct {
	TransactionID string      `json:"transactionID"`
	Operations    []Operation `json:"operations"`
}

type Operation struct {
	OperationID string `json:"operationID"`
	TypeName    string `json:"type"`
}
