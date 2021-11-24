package services

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/network"
	"github.com/stellar/go/txnbuild"
)

var accMain_pub = "GB6YYF7Q3M2TX4IGPWIJNL7HPJSD5HFVNTC4IEFYKIA7WZEPMHFFTSXM"
var accMain_sec = "SDZBCLOZS5OXTM35LX46MSPVLELLBXOUO5KHNCONIWKEJDPVCFFZ2WOM"

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) CheckTrustline(accPub string) error {

	client := horizonclient.DefaultTestNetClient

	accRequest := horizonclient.AccountsRequest{
		Asset: "GRADER:" + accMain_pub,
	}

	accPage, err := client.Accounts(accRequest)

	if err != nil {
		return err
	}

	for _, val := range accPage.Embedded.Records {
		if val.AccountID == accPub {
			return fmt.Errorf("Trustline already")
		}
	}

	return nil

}

func (s *Service) GetAllTransaction(accPub string) ([]Transaction, error) {

	client := horizonclient.DefaultTestNetClient

	accRequest := horizonclient.TransactionRequest{
		ForAccount: accPub,
	}

	page, err := client.Transactions(accRequest)

	if err != nil {
		return make([]Transaction, 0), err
	}

	transaction_list := page.Embedded.Records
	transactions := make([]Transaction, 0)

	for _, val := range transaction_list {

		request := horizonclient.OperationRequest{
			ForTransaction: val.ID,
		}

		opsPage, err := client.Operations(request)

		if err != nil {
			return make([]Transaction, 0), err
		}

		ops_list := opsPage.Embedded.Records
		ops := make([]Operation, 0)

		for _, op := range ops_list {
			ops = append(ops, Operation{
				OperationID: op.GetID(),
				TypeName:    op.GetType(),
			})
		}

		transactions = append(transactions, Transaction{
			TransactionID: val.ID,
			Operations:    ops,
		})

	}

	return transactions, nil
}

func (s *Service) Authorization(req *AuthorizationRequest) (*AuthorizationResponse, error) {

	client := horizonclient.DefaultTestNetClient

	accMainPair, err := keypair.ParseFull(accMain_sec)

	if err != nil {
		return &AuthorizationResponse{
			Status:          "Fail",
			ErrorLog:        fmt.Sprint(err),
			TransactionHash: "",
		}, err
	}

	accStudentPair, err := keypair.ParseFull(req.SecretKey)

	if err != nil {
		return &AuthorizationResponse{
			Status:          "Fail",
			ErrorLog:        fmt.Sprint(err),
			TransactionHash: "",
		}, err
	}

	accMain, err := client.AccountDetail(horizonclient.AccountRequest{
		AccountID: accMain_pub,
	})

	if err != nil {
		return &AuthorizationResponse{
			Status:          "Fail",
			ErrorLog:        fmt.Sprint(err),
			TransactionHash: "",
		}, err
	}

	accStudent, err := client.AccountDetail(horizonclient.AccountRequest{
		AccountID: req.PublicKey,
	})

	if err != nil {
		return &AuthorizationResponse{
			Status:          "Fail",
			ErrorLog:        fmt.Sprint(err),
			TransactionHash: "",
		}, err
	}

	err = s.CheckTrustline(accStudent.AccountID)

	if err != nil {
		return &AuthorizationResponse{
			Status:          "Fail",
			ErrorLog:        fmt.Sprint(err),
			TransactionHash: "",
		}, err
	}

	nativeAsset := txnbuild.NativeAsset{}
	graderAsset := txnbuild.CreditAsset{
		Code:   "GRADER",
		Issuer: accMain_pub,
	}

	graderFees := txnbuild.Payment{
		Destination: req.PublicKey,
		Amount:      "0.1",
		Asset:       &nativeAsset,
	}
	beginSponsor := txnbuild.BeginSponsoringFutureReserves{
		SourceAccount: accMain_pub,
		SponsoredID:   req.PublicKey,
	}
	changeTrust := txnbuild.ChangeTrust{
		SourceAccount: req.PublicKey,
		Line: txnbuild.ChangeTrustAssetWrapper{
			Asset: &graderAsset,
		},
	}
	endSponsor := txnbuild.EndSponsoringFutureReserves{
		SourceAccount: req.PublicKey,
	}

	graderCoin := txnbuild.Payment{
		Destination: req.PublicKey,
		Amount:      "100000",
		Asset:       &graderAsset,
	}

	ops := []txnbuild.Operation{
		&graderFees,
		&beginSponsor,
		&changeTrust,
		&endSponsor,
		&graderCoin,
	}

	sha := sha256.Sum224([]byte(req.StudentId + req.Pin))
	hash := base64.StdEncoding.EncodeToString([]byte(sha[:]))

	tx, err := txnbuild.NewTransaction(txnbuild.TransactionParams{
		SourceAccount:        &accMain,
		IncrementSequenceNum: true,
		Operations:           ops,
		BaseFee:              txnbuild.MinBaseFee,
		Timebounds:           txnbuild.NewTimeout(100),
		Memo:                 txnbuild.MemoText(hash[:28]),
	})
	if err != nil {
		return &AuthorizationResponse{
			Status:          "Fail",
			ErrorLog:        fmt.Sprint(err, hash),
			TransactionHash: "",
		}, err
	}

	tx, err = tx.Sign(network.TestNetworkPassphrase, accMainPair, accStudentPair)
	if err != nil {
		return &AuthorizationResponse{
			Status:          "Fail",
			ErrorLog:        fmt.Sprint(err),
			TransactionHash: "",
		}, err
	}

	txe, err := tx.Base64()
	if err != nil {
		return &AuthorizationResponse{
			Status:          "Fail",
			ErrorLog:        fmt.Sprint(err),
			TransactionHash: "",
		}, err
	}

	resp, err := client.SubmitTransactionXDR(txe)
	if err != nil {
		hError := err.(*horizonclient.Error)
		return &AuthorizationResponse{
			Status:          "Fail",
			ErrorLog:        fmt.Sprint(hError),
			TransactionHash: "",
		}, hError
	}

	return &AuthorizationResponse{
		Status:          "OK",
		TransactionHash: resp.Hash,
	}, nil

}
