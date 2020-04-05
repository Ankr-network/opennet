package app

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/rs/zerolog/log"
)

// IPHandler according to request args and return IP
func IPHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getIP(w, r)
	case http.MethodDelete:
		delIP(w, r)
	}
}

type getIPReq struct {
	Dev         string `json:"dev,omitempty"`
	CNI         []byte `json:"cni"`
	ContainerID string `json:"containerID"`
}

const (
	internalError = "internel server error"
)

func getIP(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error().Msg(err.Error())
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	log.Debug().Str("Request", string(body)).Msg("getIP")
	// parse request body
	gr := &getIPReq{}
	if err := json.Unmarshal(body, gr); err != nil {
		log.Error().Msg(err.Error())
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
	n := &Net{}
	if err := json.Unmarshal(gr.CNI, n); err != nil {
		log.Error().Msg(err.Error())
		return
	}
	// get ip from ip pool
	ipconfig, route, err := allocatedIP(n, gr.Dev, gr.ContainerID)
	if err != nil {
		log.Error().Msg(err.Error())
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
	result := &current.Result{}
	result.DNS = n.IPAM.DNS
	result.Routes = append(result.Routes, route)
	result.IPs = append(result.IPs, ipconfig)

	rsp, err := json.Marshal(&result)
	if err != nil {
		log.Error().Msg(err.Error())
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(rsp)
}

// DelIPReq ...
type DelIPReq struct {
	Dev         string `json:"dev"`
	ContainerID string `json:"containerID"`
}

func delIP(w http.ResponseWriter, r *http.Request) {

	req := &DelIPReq{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error().Msg(err.Error())
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	if err = json.Unmarshal(body, req); err != nil {
		log.Error().Msg(err.Error())
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}

	if err := deallocateIP(req.Dev, req.ContainerID); err != nil {
		log.Error().Msg(err.Error())
		http.Error(w, internalError, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
