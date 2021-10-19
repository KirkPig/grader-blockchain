package main

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/network"
	"github.com/stellar/go/txnbuild"
)

var accMain_pub = "GBEQ4NSJHC2VLJH3UH2HYVE6D6UVCRONXPFOZSQTOTSAY3BLYSZPIU65"
var accMain_sec = "SBKYM5IAUGQTWI4AZMG4NCVT6A3WQHLMG3NBNSOZ3JAYD6NONQD3TQUD"

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

type SubmitRequest struct {
	PublicKey string `json:"publicKey" binding:"required"`
	SecretKey string `json:"secretKey" binding:"required"`
	StudentId string `json:"studentId" binding:"required"`
	Pin       string `json:"pin" binding:"required"`
	Code      string `json:"code" binding:"required"`
}

type SubmitResponse struct {
	Status          string `json:"status" binding:"required"`
	TransactionHash string `json:"transactionHash" binding:"required"`
	ErrorLog        string `json:"errorLog",omitempty`
}

// album represents data about a record album.
type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

// albums slice to seed record album data.
var albums = []album{
	{ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
	{ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
	{ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
}

func main() {
	router := gin.Default()
	router.GET("/albums", getAlbums)
	router.GET("/albums/:id", getAlbumByID)
	router.POST("/albums", postAlbums)

	router.POST("/api/v1/authorization/new", authorization)

	router.Run("localhost:1323")
}

func submit(c *gin.Context) {

	client := horizonclient.DefaultTestNetClient

	var req SubmitRequest

	if err := c.BindJSON(&req); err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, &SubmitResponse{
			Status:          "Fail",
			ErrorLog:        fmt.Sprint(err),
			TransactionHash: "",
		})
	}

	accStudentPair, err := keypair.ParseFull(req.SecretKey)
	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, &SubmitResponse{
			Status:          "Fail",
			ErrorLog:        fmt.Sprint(err),
			TransactionHash: "",
		})
	}

	accStudent, err := client.AccountDetail(horizonclient.AccountRequest{
		AccountID: req.PublicKey,
	})

	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, &SubmitResponse{
			Status:          "Fail",
			ErrorLog:        fmt.Sprint(err),
			TransactionHash: "",
		})
	}

	// Check Authorization

	asset := txnbuild.NativeAsset{}

	paymentOp := txnbuild.Payment{
		Destination: accMain_pub,
		Amount:      string(txnbuild.MinBaseFee),
		Asset:       asset,
	}

	sha := sha256.Sum256([]byte(req.StudentId + req.Pin + req.Code))
	hash := base64.StdEncoding.EncodeToString([]byte(sha[:]))

	tx, err := txnbuild.NewTransaction(txnbuild.TransactionParams{
		SourceAccount:        &accStudent,
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
		c.IndentedJSON(http.StatusBadRequest, &SubmitResponse{
			Status:          "Fail",
			ErrorLog:        fmt.Sprint(err, hash),
			TransactionHash: "",
		})
		return
	}

	tx, err = tx.Sign(network.TestNetworkPassphrase, accStudentPair)
	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, &SubmitResponse{
			Status:          "Fail",
			ErrorLog:        fmt.Sprint(err),
			TransactionHash: "",
		})
		return
	}

	txe, err := tx.Base64()
	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, &SubmitResponse{
			Status:          "Fail",
			ErrorLog:        fmt.Sprint(err),
			TransactionHash: "",
		})
		return
	}

	resp, err := client.SubmitTransactionXDR(txe)
	if err != nil {
		hError := err.(*horizonclient.Error)
		log.Fatal("Error submitting transaction:", hError)
		c.IndentedJSON(http.StatusBadRequest, &SubmitResponse{
			Status:          "Fail",
			ErrorLog:        fmt.Sprint(hError),
			TransactionHash: "",
		})
		return
	}

	c.IndentedJSON(http.StatusOK, &SubmitResponse{
		Status:          "OK",
		TransactionHash: resp.Hash,
	})

}

func authorization(c *gin.Context) {

	client := horizonclient.DefaultTestNetClient

	var req AuthorizationRequest

	if err := c.BindJSON(&req); err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, &AuthorizationResponse{
			Status:          "Fail",
			ErrorLog:        fmt.Sprint(err),
			TransactionHash: "",
		})
		return
	}

	log.Println(req)

	accMainPair, err := keypair.ParseFull(accMain_sec)

	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, &AuthorizationResponse{
			Status:          "Fail",
			ErrorLog:        fmt.Sprint(err),
			TransactionHash: "",
		})
		return
	}

	accMain, err := client.AccountDetail(horizonclient.AccountRequest{
		AccountID: accMain_pub,
	})

	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, &AuthorizationResponse{
			Status:          "Fail",
			ErrorLog:        fmt.Sprint(err),
			TransactionHash: "",
		})
		return
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
		c.IndentedJSON(http.StatusBadRequest, &AuthorizationResponse{
			Status:          "Fail",
			ErrorLog:        fmt.Sprint(err, hash),
			TransactionHash: "",
		})
		return
	}

	tx, err = tx.Sign(network.TestNetworkPassphrase, accMainPair)
	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, &AuthorizationResponse{
			Status:          "Fail",
			ErrorLog:        fmt.Sprint(err),
			TransactionHash: "",
		})
		return
	}

	txe, err := tx.Base64()
	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusBadRequest, &AuthorizationResponse{
			Status:          "Fail",
			ErrorLog:        fmt.Sprint(err),
			TransactionHash: "",
		})
		return
	}

	resp, err := client.SubmitTransactionXDR(txe)
	if err != nil {
		hError := err.(*horizonclient.Error)
		log.Fatal("Error submitting transaction:", hError)
		c.IndentedJSON(http.StatusBadRequest, &AuthorizationResponse{
			Status:          "Fail",
			ErrorLog:        fmt.Sprint(hError),
			TransactionHash: "",
		})
		return
	}

	c.IndentedJSON(http.StatusOK, &AuthorizationResponse{
		Status:          "OK",
		TransactionHash: resp.Hash,
	})

}

// getAlbums responds with the list of all albums as JSON.
func getAlbums(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, albums)
}

// postAlbums adds an album from JSON received in the request body.
func postAlbums(c *gin.Context) {
	var newAlbum album

	// Call BindJSON to bind the received JSON to
	// newAlbum.
	if err := c.BindJSON(&newAlbum); err != nil {
		return
	}

	// Add the new album to the slice.
	albums = append(albums, newAlbum)
	c.IndentedJSON(http.StatusCreated, newAlbum)
}

// getAlbumByID locates the album whose ID value matches the id
// parameter sent by the client, then returns that album as a response.
func getAlbumByID(c *gin.Context) {
	id := c.Param("id")

	// Loop through the list of albums, looking for
	// an album whose ID value matches the parameter.
	for _, a := range albums {
		if a.ID == id {
			c.IndentedJSON(http.StatusOK, a)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
}
