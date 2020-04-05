package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os/exec"
	"strings"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/containernetworking/plugins/pkg/utils/buildversion"
	"github.com/rs/zerolog/log"
)

const (
	extSrvPath  = "/tmp/opennet.sock"
	confVersion = "0.3.1"
)

func init() {
	log.Logger = log.With().Caller().Logger()
}

func main() {
	skel.PluginMain(cmdAdd, cmdGet, cmdDel, version.All,
		buildversion.BuildString("opennet"))
}

func cmdGet(args *skel.CmdArgs) error {
	return fmt.Errorf("CNI GET method is not implemented")
}

func cmdAdd(args *skel.CmdArgs) error {
	devNo, err := getDeviceNo()
	if err != nil {
		log.Error().Msg(err.Error())
		return err
	}

	c := newHTTPClient()
	r := &InArgs{
		Dev:         devNo,
		ContainerID: args.ContainerID,
		CNI:         args.StdinData,
	}

	rb, err := json.Marshal(r)
	if err != nil {
		log.Error().Msg(err.Error())
		return err
	}

	rr, err := http.NewRequest(http.MethodGet, "http://127.0.0.1/ip", bytes.NewReader(rb))
	if err != nil {
		log.Error().Msg(err.Error())
		return err
	}

	rsp, err := c.Do(rr)
	if err != nil {
		log.Error().Msg(err.Error())
		return err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("occur unexpected event")
	}

	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		log.Error().Msg(err.Error())
		return err
	}
	cr := &current.Result{}
	if err = json.Unmarshal(body, cr); err != nil {
		log.Error().Msg(err.Error())
		return err
	}

	return types.PrintResult(cr, confVersion)
}

// DelIPReq ...
type DelIPReq struct {
	Dev         string `json:"dev"`
	ContainerID string `json:"containerID"`
}

func cmdDel(args *skel.CmdArgs) error {
	devNo, err := getDeviceNo()
	if err != nil {
		log.Error().Msg(err.Error())
		return err
	}
	req := &DelIPReq{
		Dev:         devNo,
		ContainerID: args.ContainerID,
	}
	reqData, err := json.Marshal(req)
	if err != nil {
		log.Error().Msg(err.Error())
		return err
	}

	c := newHTTPClient()
	rr, err := http.NewRequest(http.MethodDelete, "http://127.0.0.1/ip",
		bytes.NewReader(reqData))
	if err != nil {
		log.Error().Msg(err.Error())
		return err
	}
	rsp, err := c.Do(rr)
	if err != nil {
		log.Error().Msg(err.Error())
		return err
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("occur unexpected event")
	}
	return nil
}

func getDeviceNo() (string, error) {
	devNo, err := ioutil.ReadFile("/etc/cni/net.d/opennet-devno")
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(string(devNo), "\n", ""), nil
}

func getDefaultRouter() (string, error) {
	// get default router dev
	out, err := exec.Command("/bin/bash", "-c",
		"ip route | grep '^default' | awk '{printf $5}'").Output()
	if err != nil {
		return "", err
	}
	nio, err := net.InterfaceByName(fmt.Sprintf("%s", out))
	if err != nil {
		return "", err
	}

	addrs, err := nio.Addrs()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s", strings.Split(addrs[0].String(), "/")[0]), nil
}

// InArgs request ip
type InArgs struct {
	Dev         string `json:"dev"`
	CNI         []byte `json:"cni"`
	ContainerID string `json:"containerID"`
}

func newHTTPClient() *http.Client {
	tr := &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			return net.Dial("unix", extSrvPath)
		},
	}
	return &http.Client{Transport: tr}
}
