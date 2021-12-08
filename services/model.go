package services

type AuthorizationRequest struct {
	PublicKey string `json:"publicKey" binding:"required"`
	SecretKey string `json:"secretKey" binding:"required"`
	StudentId string `json:"studentId" binding:"required"`
	Pin       string `json:"pin" binding:"required"`
}

type SentCodeRequest struct {
	PublicKey string `json:"publicKey"`
	StudentId string `json:"studentId"`
	Pin       string `json:"pin"`
	Code      string `json:"code"`
}

type CheckCodeRequest struct {
	PublicKey string `json:"publicKey"`
	StudentId string `json:"studentId"`
	Pin       string `json:"pin"`
	Code      string `json:"code"`
}

type Response struct {
	Status          string `json:"status"`
	TransactionHash string `json:"transactionHash"`
	ErrorLog        string `json:"errorLog"`
}

type Transaction struct {
	TransactionID string      `json:"transactionID"`
	Operations    []Operation `json:"operations"`
}

type Operation struct {
	OperationID string `json:"operationID"`
	TypeName    string `json:"type"`
}
