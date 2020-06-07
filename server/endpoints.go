package main

import (
	"encoding/json"
	"net/http"
)

func instructOn(w http.ResponseWriter, r *http.Request) {
	instruction := CompInstruction{
		Type: "power_on",
	}
	w.WriteHeader(http.StatusOK)
	torrent <- instruction
	json.NewEncoder(w).Encode(OK200)
}

func instructOff(w http.ResponseWriter, r *http.Request) {
	instruction := CompInstruction{
		Type: "power_off",
	}
	w.WriteHeader(http.StatusOK)
	torrent <- instruction
	json.NewEncoder(w).Encode(OK200)

}

func restartPulse(w http.ResponseWriter, r *http.Request) {
	instruction := CompInstruction{
		Type: "restart_pulse",
	}
	w.WriteHeader(http.StatusOK)
	torrent <- instruction
	json.NewEncoder(w).Encode(OK200)
}
