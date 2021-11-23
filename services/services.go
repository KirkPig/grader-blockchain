package services

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/network"
	"github.com/stellar/go/txnbuild"
)

var accMain_pub = "GBEQ4NSJHC2VLJH3UH2HYVE6D6UVCRONXPFOZSQTOTSAY3BLYSZPIU65"
var accMain_sec = "SBKYM5IAUGQTWI4AZMG4NCVT6A3WQHLMG3NBNSOZ3JAYD6NONQD3TQUD"

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Authorization(req *AuthorizationRequest) *AuthorizationResponse {

	client := horizonclient.DefaultTestNetClient

	accMainPair, err := keypair.ParseFull(accMain_sec)

	if err != nil {
		log.Println(err)
		return &AuthorizationResponse{
			Status:          "Fail",
			ErrorLog:        fmt.Sprint(err),
			TransactionHash: "",
		}
	}

	accMain, err := client.AccountDetail(horizonclient.AccountRequest{
		AccountID: accMain_pub,
	})

	if err != nil {
		log.Println(err)
		return &AuthorizationResponse{
			Status:          "Fail",
			ErrorLog:        fmt.Sprint(err),
			TransactionHash: "",
		}
	}

	/*


			accStudent, err := client.AccountDetail(horizonclient.AccountRequest{
				AccountID: req.publicKey,
			})

		if err != nil {
			log.Println(err)
			return
		}*/

	asset := txnbuild.NativeAsset{}

	paymentOp := txnbuild.Payment{
		Destination: req.PublicKey,
		Amount:      "1",
		Asset:       asset,
	}

	sha := sha256.Sum224([]byte(req.StudentId + req.Pin))
	hash := base64.StdEncoding.EncodeToString([]byte(sha[:]))

	tx, err := txnbuild.NewTransaction(txnbuild.TransactionParams{
		SourceAccount:        &accMain,
		IncrementSequenceNum: true,
		Operations: []txnbuild.Operation{
			&paymentOp,
		},
		BaseFee:    txnbuild.MinBaseFee,
		Timebounds: txnbuild.NewTimeout(100),
		Memo:       txnbuild.MemoText(hash),
	})
	if err != nil {
		log.Println(err)
		return &AuthorizationResponse{
			Status:          "Fail",
			ErrorLog:        fmt.Sprint(err, hash),
			TransactionHash: "",
		}
	}

	tx, err = tx.Sign(network.TestNetworkPassphrase, accMainPair)
	if err != nil {
		log.Println(err)
		return &AuthorizationResponse{
			Status:          "Fail",
			ErrorLog:        fmt.Sprint(err),
			TransactionHash: "",
		}
	}

	txe, err := tx.Base64()
	if err != nil {
		log.Println(err)
		return &AuthorizationResponse{
			Status:          "Fail",
			ErrorLog:        fmt.Sprint(err),
			TransactionHash: "",
		}
	}

	resp, err := client.SubmitTransactionXDR(txe)
	if err != nil {
		hError := err.(*horizonclient.Error)
		log.Fatal("Error submitting transaction:", hError)
		return &AuthorizationResponse{
			Status:          "Fail",
			ErrorLog:        fmt.Sprint(hError),
			TransactionHash: "",
		}
	}

	return &AuthorizationResponse{
		Status:          "OK",
		TransactionHash: resp.Hash,
	}

}
