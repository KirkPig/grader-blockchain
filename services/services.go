package services

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/network"
	"github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/txnbuild"
)

var accMain_pub = "GATNBMBQEPZ32HZQU4RMAHQJVITXBZS4OVQKFK7BWVZSRALYB3VKEVG4"
var accMain_sec = "SA3XV35IDKTDROY5UERQIJUSJ6QQCGIRKAULECV25HUQPG3J7UWPK47X"

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) GetBalances() ([]horizon.Balance, error) {

	client := horizonclient.DefaultTestNetClient

	accRequest := horizonclient.AccountRequest{
		AccountID: accMain_pub,
	}

	account, err := client.AccountDetail(accRequest)

	if err != nil {
		return make([]horizon.Balance, 0), err
	} else {
		return account.Balances, err
	}

}

func (s *Service) RemoveAllTrustline() (string, error) {
	client := horizonclient.DefaultTestNetClient

	accMainPair, err := keypair.ParseFull(accMain_sec)

	if err != nil {
		return "", err
	}

	accMain, err := client.AccountDetail(horizonclient.AccountRequest{
		AccountID: accMain_pub,
	})

	bal, err := s.GetBalances()

	if err != nil {
		return "", err
	}

	ops := make([]txnbuild.Operation, 0)

	for _, b := range bal {

		asset := txnbuild.CreditAsset{
			Code:   "GRADER",
			Issuer: b.Issuer,
		}

		if b.Type != "native" {

			if b.Balance != "0.0000000" {
				ops = append(ops, &txnbuild.Payment{
					Destination: b.Issuer,
					Asset:       asset,
					Amount:      b.Balance,
				})
			}

			ops = append(ops, &txnbuild.ChangeTrust{
				Line: txnbuild.ChangeTrustAssetWrapper{
					Asset: &asset,
				},
				Limit: "0",
			})

		}

	}

	tx, err := txnbuild.NewTransaction(txnbuild.TransactionParams{
		SourceAccount:        &accMain,
		IncrementSequenceNum: true,
		Operations:           ops,
		BaseFee:              txnbuild.MinBaseFee,
		Timebounds:           txnbuild.NewTimeout(100),
	})
	if err != nil {
		return "", err
	}

	tx, err = tx.Sign(network.TestNetworkPassphrase, accMainPair)
	if err != nil {
		return "", err
	}

	txe, err := tx.Base64()
	if err != nil {
		return "", err
	}

	resp, err := client.SubmitTransactionXDR(txe)
	if err != nil {
		hError := err.(*horizonclient.Error)
		return "", hError
	}

	return resp.Hash, nil
}

func (s *Service) RemoveTrustlines(accPub_list []string) (string, error) {

	client := horizonclient.DefaultTestNetClient

	accMainPair, err := keypair.ParseFull(accMain_sec)

	if err != nil {
		return "", err
	}

	accMain, err := client.AccountDetail(horizonclient.AccountRequest{
		AccountID: accMain_pub,
	})

	bal, err := s.GetBalances()

	if err != nil {
		return "", err
	}

	ops := make([]txnbuild.Operation, 0)

	for _, val := range accPub_list {

		chk := false

		for _, b := range bal {
			if b.Issuer == val {
				chk = true

				asset := txnbuild.CreditAsset{
					Code:   "GRADER",
					Issuer: b.Issuer,
				}

				if b.Type != "native" {
					if b.Balance != "0.0000000" {
						ops = append(ops, &txnbuild.Payment{
							Destination: b.Issuer,
							Asset:       asset,
							Amount:      b.Balance,
						})
					}

					ops = append(ops, &txnbuild.ChangeTrust{
						Line: txnbuild.ChangeTrustAssetWrapper{
							Asset: &asset,
						},
						Limit: "0",
					})
				}

				break
			}
		}

		if !chk {
			return "", fmt.Errorf("Some accPub are not in Trustline")
		}

	}

	tx, err := txnbuild.NewTransaction(txnbuild.TransactionParams{
		SourceAccount:        &accMain,
		IncrementSequenceNum: true,
		Operations:           ops,
		BaseFee:              txnbuild.MinBaseFee,
		Timebounds:           txnbuild.NewTimeout(100),
	})
	if err != nil {
		return "", err
	}

	tx, err = tx.Sign(network.TestNetworkPassphrase, accMainPair)
	if err != nil {
		return "", err
	}

	txe, err := tx.Base64()
	if err != nil {
		return "", err
	}

	resp, err := client.SubmitTransactionXDR(txe)
	if err != nil {
		hError := err.(*horizonclient.Error)
		return "", hError
	}

	return resp.Hash, nil

}

func (s *Service) CheckTrustline(accPub string) (bool, error) {

	client := horizonclient.DefaultTestNetClient

	accRequest := horizonclient.AccountsRequest{
		Asset: "GRADER:" + accPub,
	}

	accPage, err := client.Accounts(accRequest)

	if err != nil {
		return false, err
	}

	for _, val := range accPage.Embedded.Records {
		if val.AccountID == accMain_pub {
			return true, fmt.Errorf("Trustline already")
		}
	}

	return false, nil

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

func (s *Service) GetHashToken(studentID string, pin string, accPub string) string {
	sha := sha256.Sum224([]byte(studentID + pin + accPub))
	return base64.StdEncoding.EncodeToString([]byte(sha[:]))
}

func (s *Service) GetAllMemo() ([]string, error) {

	client := horizonclient.DefaultTestNetClient

	accRequest := horizonclient.TransactionRequest{
		ForAccount: accMain_pub,
	}

	page, err := client.Transactions(accRequest)

	if err != nil {
		return make([]string, 0), err
	}

	transaction_list := page.Embedded.Records
	memo_list := make([]string, 0)

	for _, val := range transaction_list {
		memo_list = append(memo_list, val.Memo)
	}

	return memo_list, nil

}

func (s *Service) MemoCheck(auth string, memo_list []string) bool {

	for _, val := range memo_list {
		if val == auth {
			return true
		}
	}

	return false
}

func (s *Service) CheckCode(req CheckCodeRequest) (string, error) {

	memo_list, err := s.GetAllMemo()

	if err != nil {
		return "", err
	}

	auth := s.GetHashToken(req.StudentId, req.Pin, req.PublicKey)
	authCheck := s.MemoCheck(auth[:28], memo_list)

	if !authCheck {
		return "", fmt.Errorf("Can't Find Any of your authorization")
	}

	sha := sha256.Sum224([]byte(auth + req.Code))
	hash := base64.StdEncoding.EncodeToString([]byte(sha[:]))

	codeCheck := s.MemoCheck(hash[:28], memo_list)

	if !codeCheck {
		return "", fmt.Errorf("It's not your code")
	}

	return "", nil
}

func (s *Service) SentCode(req SentCodeRequest) (string, error) {

	client := horizonclient.DefaultTestNetClient

	accMainPair, err := keypair.ParseFull(accMain_sec)

	if err != nil {
		return "", err
	}

	accMain, err := client.AccountDetail(horizonclient.AccountRequest{
		AccountID: accMain_pub,
	})

	if err != nil {
		return "", err
	}

	memo_list, err := s.GetAllMemo()

	if err != nil {
		return "", err
	}

	auth := s.GetHashToken(req.StudentId, req.Pin, req.PublicKey)
	authCheck := s.MemoCheck(auth[:28], memo_list)

	if !authCheck {
		return "", fmt.Errorf("Can't Find Any of your authorization")
	}

	sha := sha256.Sum224([]byte(auth + req.Code))
	hash := base64.StdEncoding.EncodeToString([]byte(sha[:]))

	graderAsset := txnbuild.CreditAsset{
		Code:   "GRADER",
		Issuer: req.PublicKey,
	}

	graderCoin := txnbuild.Payment{
		Destination: req.PublicKey,
		Amount:      "1",
		Asset:       &graderAsset,
	}

	ops := []txnbuild.Operation{
		&graderCoin,
	}

	tx, err := txnbuild.NewTransaction(txnbuild.TransactionParams{
		SourceAccount:        &accMain,
		IncrementSequenceNum: true,
		Operations:           ops,
		BaseFee:              txnbuild.MinBaseFee,
		Timebounds:           txnbuild.NewTimeout(100),
		Memo:                 txnbuild.MemoText(hash[:28]),
	})
	if err != nil {
		return "", err
	}

	tx, err = tx.Sign(network.TestNetworkPassphrase, accMainPair)
	if err != nil {
		return "", err
	}

	txe, err := tx.Base64()
	if err != nil {
		return "", err
	}

	resp, err := client.SubmitTransactionXDR(txe)
	if err != nil {
		hError := err.(*horizonclient.Error)
		return "", hError
	}

	return resp.Hash, nil

}

func (s *Service) Authorization(req *AuthorizationRequest) (string, error) {

	client := horizonclient.DefaultTestNetClient

	accMainPair, err := keypair.ParseFull(accMain_sec)

	if err != nil {
		return "", err
	}

	accStudentPair, err := keypair.ParseFull(req.SecretKey)

	if err != nil {
		return "", err
	}

	accMain, err := client.AccountDetail(horizonclient.AccountRequest{
		AccountID: accMain_pub,
	})

	if err != nil {
		return "", err
	}

	accStudent, err := client.AccountDetail(horizonclient.AccountRequest{
		AccountID: req.PublicKey,
	})

	if err != nil {
		return "", err
	}

	/* Check Duplicate Trustline */

	_, err = s.CheckTrustline(accStudent.AccountID)

	if err != nil {
		return "", err
	}

	graderAsset := txnbuild.CreditAsset{
		Code:   "GRADER",
		Issuer: req.PublicKey,
	}

	beginSponsor := txnbuild.BeginSponsoringFutureReserves{
		SourceAccount: accMain_pub,
		SponsoredID:   req.PublicKey,
	}
	changeTrust := txnbuild.ChangeTrust{
		Line: txnbuild.ChangeTrustAssetWrapper{
			Asset: &graderAsset,
		},
	}
	endSponsor := txnbuild.EndSponsoringFutureReserves{
		SourceAccount: req.PublicKey,
	}

	graderCoin := txnbuild.Payment{
		SourceAccount: req.PublicKey,
		Destination:   accMain_pub,
		Amount:        "100000",
		Asset:         &graderAsset,
	}

	ops := []txnbuild.Operation{
		&beginSponsor,
		&changeTrust,
		&endSponsor,
		&graderCoin,
	}

	hash := s.GetHashToken(req.StudentId, req.Pin, req.PublicKey)

	tx, err := txnbuild.NewTransaction(txnbuild.TransactionParams{
		SourceAccount:        &accMain,
		IncrementSequenceNum: true,
		Operations:           ops,
		BaseFee:              txnbuild.MinBaseFee,
		Timebounds:           txnbuild.NewTimeout(100),
		Memo:                 txnbuild.MemoText(hash[:28]),
	})
	if err != nil {
		return "", err
	}
	tx, err = tx.Sign(network.TestNetworkPassphrase, accMainPair)
	tx, err = tx.Sign(network.TestNetworkPassphrase, accStudentPair)
	if err != nil {
		return "", err
	}

	txe, err := tx.Base64()
	if err != nil {
		return "", err
	}

	resp, err := client.SubmitTransactionXDR(txe)
	if err != nil {
		hError := err.(*horizonclient.Error)
		return "", hError
	}

	return resp.Hash, nil

}
