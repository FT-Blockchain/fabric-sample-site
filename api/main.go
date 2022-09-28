/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"encoding/json"
	"path/filepath"
	"net/http"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

type Asset struct {
	AssetID string `json:"asset_id"`
	Owner string `json:"owner"`
	Colour string `json:"colour"`
	Size string `json:"size"`
	AppraisedValue string `json:"appraised_value"`
}

type PostTransaction struct {
	AssetID string `json:"asset_id"`
	Owner string `json:"owner"`
}

type PostAsset struct {
	Id string	`json:"id"`
}

type walletHandler struct {
	wallet *gateway.Wallet
	contract *gateway.Contract
}

func (wh *walletHandler) CreateAsset(w http.ResponseWriter, req *http.Request) {
	setupCORS(&w, req)
    if (*req).Method == "OPTIONS" {
        return
    }

	if req.Method == "POST" {

		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(w, "Error reading request body",
				http.StatusInternalServerError)
		}

		asset := Asset{}
		json.Unmarshal(body, &asset)

		exists := checkIfAssetExists(wh.contract, asset.AssetID)

		log.Println(exists)

		if exists {
			w.Write([]byte("error asset already exists"))
			return
		}

		log.Println("--> Submit Transaction: CreateAsset, creates new asset with ID, color, owner, size, and appraisedValue arguments")
		result, err := wh.contract.SubmitTransaction("CreateAsset", asset.AssetID, asset.Colour, asset.Size, asset.Owner, asset.AppraisedValue)
		if err != nil {
			log.Fatalf("Failed to Submit transaction: %v", err)
		}

		w.Write(result)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func (wh *walletHandler) StartTransaction(w http.ResponseWriter, req *http.Request) {
	setupCORS(&w, req)
    if (*req).Method == "OPTIONS" {
        return
    }

	if req.Method == "POST" {

		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(w, "Error reading request body",
				http.StatusInternalServerError)
		}

		transaction := PostTransaction{}
		json.Unmarshal(body, &transaction)

		exists := checkIfAssetExists(wh.contract, transaction.AssetID)

		if !exists {
			w.Write([]byte("error asset does not exists"))
			return
		}

		log.Println("--> Submit Transaction: TransferAsset asset1, transfer to new owner of Tom")
		result, err := wh.contract.SubmitTransaction("TransferAsset", transaction.AssetID, transaction.Owner)
		if err != nil {
			log.Fatalf("Failed to Submit transaction: %v", err)
		}

		w.Write(result)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func (wh *walletHandler) GetAllAssets(w http.ResponseWriter, req *http.Request) {
	setupCORS(&w, req)
    if (*req).Method == "OPTIONS" {
        return
    }

	log.Println("--> Evaluate Transaction: GetAllAssets, function returns all the current assets on the ledger")
	result, err := wh.contract.EvaluateTransaction("GetAllAssets")
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %v", err)
	}
	log.Println(string(result))

	w.Write(result)
}

func (wh *walletHandler) GetSingleAsset(w http.ResponseWriter, req *http.Request) {
	setupCORS(&w, req)
    if (*req).Method == "OPTIONS" {
        return
    }

	if req.Method == "POST" {

		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(w, "Error reading request body",
				http.StatusInternalServerError)
		}

		asset := PostAsset{}
		json.Unmarshal(body, &asset)

		exists := checkIfAssetExists(wh.contract, asset.Id)

		log.Println(exists)

		if !exists {
			w.Write([]byte("error asset does not exists"))
			return
		}

		log.Println("--> Evaluate Transaction: ReadAsset, function returns an asset with a given assetID")
		result, err := wh.contract.EvaluateTransaction("ReadAsset", asset.Id)
		if err != nil {
			log.Fatalf("Failed to evaluate transaction: %v\n", err)
		}
		log.Println(string(result))

		w.Write(result)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func main() {
	log.Println("============ application-golang starts ============")

	err := os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
	if err != nil {
		log.Fatalf("Error setting DISCOVERY_AS_LOCALHOST environemnt variable: %v", err)
	}

	wallet, err := gateway.NewFileSystemWallet("wallet")
	if err != nil {
		log.Fatalf("Failed to create wallet: %v", err)
	}

	if !wallet.Exists("appUser") {
		err = populateWallet(wallet)
		if err != nil {
			log.Fatalf("Failed to populate wallet contents: %v", err)
		}
	}

	ccpPath := filepath.Join(
		"connection",
		"connection-org1.yaml",
	)

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		gateway.WithIdentity(wallet, "appUser"),
	)

	if err != nil {
		log.Fatalf("Failed to connect to gateway: %v", err)
	}

	defer gw.Close()

	network, err := gw.GetNetwork("mychannel")

	if err != nil {
		log.Fatalf("Failed to get network: %v", err)
	}

	contract := network.GetContract("basic")

	log.Println("--> Submit Transaction: InitLedger, function creates the initial set of assets on the ledger")
	result, err := contract.SubmitTransaction("InitLedger")
	if err != nil {
		log.Fatalf("Failed to Submit transaction: %v", err)
	}
	log.Println(string(result))

	wHandler := walletHandler{
		wallet: wallet,
		contract: contract,
	}

	http.HandleFunc("/create-asset", wHandler.CreateAsset)
	http.HandleFunc("/transaction", wHandler.StartTransaction)
	http.HandleFunc("/assets", wHandler.GetAllAssets)
	http.HandleFunc("/asset", wHandler.GetSingleAsset)
	http.ListenAndServe(":8090", nil)
}

func populateWallet(wallet *gateway.Wallet) error {
	log.Println("============ Populating wallet ============")
	credPath := filepath.Join(
		"user",
	)

	certPath := filepath.Join(credPath, "signcerts", "cert.pem")
	// read the certificate pem
	cert, err := ioutil.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return err
	}

	keyDir := filepath.Join(credPath, "keystore")
	// there's a single file in this dir containing the private key
	files, err := ioutil.ReadDir(keyDir)
	if err != nil {
		return err
	}
	if len(files) != 1 {
		return fmt.Errorf("keystore folder should have contain one file")
	}
	keyPath := filepath.Join(keyDir, files[0].Name())
	key, err := ioutil.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return err
	}

	identity := gateway.NewX509Identity("Org1MSP", string(cert), string(key))

	return wallet.Put("appUser", identity)
}

func checkIfAssetExists(contract *gateway.Contract, asset string) bool{
	log.Println("--> Evaluate Transaction: AssetExists, function returns 'true' if an asset with given assetID exist")
	result, _ := contract.EvaluateTransaction("AssetExists", asset)
	log.Println(string(result))
	
	if string(result) == "true"{
		return true
	}

	return false
}

func setupCORS(w *http.ResponseWriter, req *http.Request) {
    (*w).Header().Set("Access-Control-Allow-Origin", "*")
    (*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
    (*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}


