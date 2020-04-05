package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/rs/zerolog/log"
)

const (
	rootPath       = "/opennet/ip"
	lockPath       = "/opennet/lock"
	requestTimeout = 30 * time.Second
)

// IPSet ...
type IPSet map[string]string

func (i IPSet) reverse() map[string]struct{} {
	m := make(map[string]struct{}, len(i))
	for _, v := range i {
		m[v] = struct{}{}
	}
	return m
}

func allocatedIP(nc *Net, dev, containerID string) (*current.IPConfig, *types.Route, error) {
	var ipr, gateway net.IP
	var netmask net.IPMask
	// lock here
	session, err := concurrency.NewSession(store.c)
	if err != nil {
		return nil, nil, err
	}
	defer session.Close()
	mux := concurrency.NewMutex(session, fmt.Sprintf("%s/%s", lockPath, dev))
	if err := mux.Lock(context.TODO()); err != nil {
		return nil, nil, err
	}
	defer func() {
		if err := mux.Unlock(context.TODO()); err != nil {
			log.Error().Msg(err.Error())
		}
	}()

	// do allocate ip actions
	kvc := clientv3.NewKV(store.c)
	ipStorePath := fmt.Sprintf("%s/%s", rootPath, dev)
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	rsp, err := kvc.Get(ctx, ipStorePath)
	cancel()
	if err != nil {
		log.Error().Msg(err.Error())
		return nil, nil, err
	}

	ipset := make(IPSet)
	if len(rsp.Kvs) > 0 {
		if err := json.Unmarshal(rsp.Kvs[0].Value, &ipset); err != nil {
			log.Error().Msg(err.Error())
			return nil, nil, err
		}
	}

	// look for appropriate ip
	rip := ipset.reverse()
	round := len(nc.IPAM.Range)
	for i := 0; i < round; i++ {
		log.Debug().Str("IDev", nc.IPAM.Range[i].Dev).Str("Dev", dev).Msg("ip dev")
		// if not specify device, then hop it
		if nc.IPAM.Range[i].Dev != dev {
			continue
		}
		ipBegin, _ := IP2long(nc.IPAM.Range[i].Start)
		ipEnd, _ := IP2long(nc.IPAM.Range[i].End)
		for ipBegin <= ipEnd {
			// if the ip isn't allocated, then allocate ip
			if _, ok := rip[Long2ip(ipBegin)]; !ok {
				ipr = Long2IPNet(ipBegin)
				_, ipnet, _ := net.ParseCIDR(nc.IPAM.Range[i].Subnet)
				netmask = ipnet.Mask
				gateway = net.ParseIP(nc.IPAM.Range[i].GW)
				goto STORE_IP
			}
			ipBegin++
		}
	}
STORE_IP:
	if ipr.String() == "" {
		return nil, nil, fmt.Errorf("no ip")
	}

	ipset[containerID] = ipr.String()

	data, err := json.Marshal(&ipset)
	if err != nil {
		log.Error().Msg(err.Error())
		return nil, nil, err
	}

	ctx, cancel = context.WithTimeout(context.Background(), requestTimeout)
	_, err = kvc.Put(ctx, ipStorePath, string(data))
	cancel()
	if err != nil {
		log.Error().Msg(err.Error())
		return nil, nil, err
	}

	return &current.IPConfig{
			Version: "4",
			Address: net.IPNet{IP: ipr, Mask: netmask},
			Gateway: gateway,
		}, &types.Route{
			Dst: net.IPNet{IP: net.ParseIP("0.0.0.0"),
				Mask: net.IPv4Mask(0, 0, 0, 0)},
			GW: gateway,
		}, nil

}

func deallocateIP(dev, containerID string) error {
	// lock here
	session, err := concurrency.NewSession(store.c)
	if err != nil {
		return err
	}
	defer session.Close()
	mux := concurrency.NewMutex(session, fmt.Sprintf("%s/%s", lockPath, dev))
	if err := mux.Lock(context.TODO()); err != nil {
		return err
	}
	defer func() {
		if err := mux.Unlock(context.TODO()); err != nil {
			log.Error().Msg(err.Error())
		}
	}()

	// recycle ip
	ipStorePath := fmt.Sprintf("%s/%s", rootPath, dev)
	kvc := clientv3.NewKV(store.c)
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	rsp, err := kvc.Get(ctx, ipStorePath)
	cancel()
	if err != nil {
		log.Error().Msg(err.Error())
		return err
	}
	ipset := make(IPSet)
	ipSetSrv := make(IPSet)
	if err := json.Unmarshal(rsp.Kvs[0].Value, &ipset); err != nil {
		log.Error().Msg(err.Error())
		return err
	}

	// recycle the ip and restore into etcd
	for k, v := range ipset {
		if k != containerID {
			ipSetSrv[k] = v
		}
	}

	data, err := json.Marshal(&ipSetSrv)
	if err != nil {
		log.Error().Msg(err.Error())
		return err
	}

	ctx, cancel = context.WithTimeout(context.Background(), requestTimeout)
	_, err = kvc.Put(ctx, ipStorePath, string(data))
	cancel()
	if err != nil {
		log.Error().Msg(err.Error())
		return err
	}

	return nil
}
