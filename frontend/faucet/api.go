package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/threefoldtech/rivine/types"
)

func (f *faucet) requestCoins(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	body := struct {
		Address types.UnlockHash `json:"address"`
	}{}

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("[DEBUG] Requesting coins (%s) through API\n", body.Address.String())

	txID, err := dripCoins(body.Address, f.coinsToGive)
	if err != nil {
		log.Println("[ERROR] Failed to drip coins:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct {
		TxID types.TransactionID `json:"txid"`
	}{TxID: txID})
}

func (f *faucet) requestAuthorization(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	body := struct {
		Address types.UnlockHash `json:"address"`
	}{}

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("[DEBUG] Requesting address authorization (%s) through API\n", body.Address.String())

	txID, err := updateAddressAuthorization(body.Address, true)
	if err != nil {
		log.Println("[ERROR] Failed to authorize address:", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct {
		TxID types.TransactionID `json:"txid"`
	}{TxID: txID})
}

func (f *faucet) requestDeauthorization(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	body := struct {
		Address types.UnlockHash `json:"address"`
	}{}

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("[DEBUG] Requesting address deauthorization (%s) through API\n", body.Address.String())

	txID, err := updateAddressAuthorization(body.Address, false)
	if err != nil {
		log.Println("[ERROR] Failed to deauthorize address:", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct {
		TxID types.TransactionID `json:"txid"`
	}{TxID: txID})
}
