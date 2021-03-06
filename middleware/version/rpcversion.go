/**
 *  author: lim
 *  data  : 18-4-10 下午10:16
 */

package version

import (
	"net/rpc"
	"sync"
	"errors"
)

const(
	DIALERR = iota
	USEING
	FREE
	CLOSED

)

var GET_VERSION_CONN_FAILED error = errors.New("GET VERSION CONN FAILED")
var RELEASE_FAILED error = errors.New("RELEASE USED VERSION FAILED ")
var lock sync.Mutex

type Cli struct {
	cli *rpc.Client
	idx int
}

func (cli *Cli) Close() {
	pool.used[cli.idx] = FREE
}

type Pool struct {
	clis []*Cli
	used []uint8
	mu   sync.RWMutex
}


var pool Pool
func NewRpcPool(size int, addr string) {
	pool = Pool{}
	pool.clis = make([]*Cli, size)
	pool.used = make([]uint8, size)

	for idx := 0; idx < size; idx += 1 {
		cli, err := rpc.DialHTTP("tcp", addr)
		pool.clis[idx] = &Cli{cli:cli, idx:idx}
		if err != nil {
			pool.used[idx] = DIALERR
		} else {
			pool.used[idx] = FREE
		}
	}
}

func (p *Pool) getConn() *Cli {
	p.mu.Lock()
	defer p.mu.Unlock()
	for idx := 0; idx < len(p.used); idx += 1 {
		if p.used[idx] == FREE {
			p.used[idx] = USEING
			return p.clis[idx]
		}
	}
	return nil
}


func (p *Pool) Close() {
	for idx, cli := range p.clis {
		cli.cli.Close()
		p.used[idx] = CLOSED
	}
}


func NextVersion() ([]byte, error) {
	cli := pool.getConn()
	if cli == nil {
		return nil, GET_VERSION_CONN_FAILED
	}
	defer cli.Close()

	var nextVer []byte
	err := cli.cli.Call("VSeq.NextV", uint8(0), &nextVer)
	if err != nil {
		return nil, err
	}

	return nextVer, nil
}

func ReleaseVersion(version []byte) error {
	cli := pool.getConn()
	if cli == nil {
		return GET_VERSION_CONN_FAILED
	}
	defer cli.Close()

	var ret bool
	err := cli.cli.Call("VSeq.Release", version, &ret)
	if err != nil {
		return err
	}
	if ret {
		return nil
	}else {
		return RELEASE_FAILED
	}
}

func VersionsInUse() ([][]byte, error ){
	cli := pool.getConn()
	if cli == nil {
		return nil, GET_VERSION_CONN_FAILED
	}
	defer cli.Close()

	var vInuse [][]byte
	err := cli.cli.Call("VSeq.VInUser", uint8(0), &vInuse)
	if err != nil {
		panic(err)
	}
	return vInuse, nil
}