package services

type AuthorizationRequest struct {
	PublicKey string `json:"publicKey" binding:"required"`
	StudentId string `json:"studentId" binding:"required"`
	Pin       string `json:"pin" binding:"required"`
}

type AuthorizationResponse struct {
	Status          string `json:"status" binding:"required"`
	TransactionHash string `json:"transactionHash" binding:"required"`
	ErrorLog        string `json:"errorLog",omitempty binding:"required"`
}
