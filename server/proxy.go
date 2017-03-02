package server

import (
	"fmt"
	"github.com/lycying/rstore"
	"github.com/lycying/rstore/codec"
	"github.com/lycying/rstore/redisx"
	"github.com/lycying/rstore/route"
	"strconv"
	"strings"
)

// proxyFunc recive a client request and return an new response
// the the mut framework send it to the client
type proxyFunc func(*codec.Request) *codec.Response

// Proxy hold gdt(function global descriptor table)
type Proxy struct {
	gdt    map[string]proxyFunc
	router *route.Router
}

// NewProxy make new redis proxy to handle request
func newProxy() *Proxy {
	proxy := &Proxy{}
	proxy.gdt = map[string]proxyFunc{
		"GET":                proxy.get,
		"SET":                proxy.set,
		"INCR":               proxy.incr,
		"DECR":               proxy.decr,
		"INCRBY":             proxy.incrby,
		"DECRBY":             proxy.decrby,
		"HSET":               proxy.hset,
		"HMSET":              proxy.hmset,
		"HGET":               proxy.hget,
		"HDEL":               proxy.hdel,
		"HLEN":               proxy.hlen,
		"HMGET":              proxy.hmget,
		"HKEYS":              proxy.hkeys,
		"HVALS":              proxy.hvals,
		"HINCRBY":            proxy.hincrby,
		"HGETALL":            proxy.hgetall,
		"HEXISTS":            proxy.hexists,
		"ZADD":               proxy.zadd,
		"ZSCORE":             proxy.zscore,
		"ZREM":               proxy.zrem,
		"ZCARD":              proxy.zcard,
		"ZCOUNT":             proxy.zcount,
		"ZRANK":              proxy.zrank,
		"ZRANGE":             proxy.zrange,
		"ZRANGEBYSCORE":      proxy.zrangebyscore,
		"ZREMRANGEBYSCORE":   proxy.zremrangebyscore,
		"ZREVRANGEWITHSCORE": proxy.proxyZrevRangeWithScore,
		"SADD":               proxy.proxySadd,
		"SCARD":              proxy.proxyScard,
		"SISMEMBER":          proxy.proxySisMember,
		"SMEMBERS":           proxy.proxySmembers,
		"SREM":               proxy.proxySrem,
	}

	proxy.router = route.NewRouter()
	return proxy
}

func (proxy *Proxy) doRouter(key string) (redisx.Redis, error) {
	db, err := proxy.router.GetDBUnit(key)
	if err != nil {
		return nil, err
	}
	return db.Backend, nil
}

func (proxy *Proxy) invoke(req *codec.Request) *codec.Response {
	cmd := strings.ToUpper(req.C)

	if f, ok := proxy.gdt[cmd]; ok {
		return f(req)
	}

	resp := codec.NewResponse()
	resp.WriteErrorString(fmt.Sprintf("rstore: unknown command '%s'", req.C))

	return resp
}

func (proxy *Proxy) get(req *codec.Request) *codec.Response {
	resp := codec.NewResponse()

	if req.ParamsLen() != 1 {
		resp.WriteError(rstore.WrongReqArgsNumber)
		return resp
	}

	k := req.P[0]

	store, err := proxy.doRouter(k)
	if err != nil {
		resp.WriteError(err)
		return resp
	}
	v, err := store.GET(k)
	if err != nil {
		if err == rstore.KeyIsNilError {
			resp.WriteNil()
		} else {
			resp.WriteError(err)
		}
	} else {
		resp.WriteString(v)
	}
	return resp
}

func (proxy *Proxy) set(req *codec.Request) *codec.Response {
	resp := codec.NewResponse()

	if req.ParamsLen() != 2 {
		resp.WriteError(rstore.WrongReqArgsNumber)
		return resp
	}

	k, v := req.P[0], req.P[1]

	store, err := proxy.doRouter(k)
	if err != nil {
		resp.WriteError(err)
		return resp
	}
	err = store.SET(k, v)
	if err == nil {
		resp.WriteOK()
	} else {
		resp.WriteError(err)
	}
	return resp
}

func (proxy *Proxy) incr(req *codec.Request) *codec.Response {
	resp := codec.NewResponse()

	if req.ParamsLen() != 1 {
		resp.WriteError(rstore.WrongReqArgsNumber)
		return resp
	}

	k := req.P[0]

	store, err := proxy.doRouter(k)
	if err != nil {
		resp.WriteError(err)
		return resp
	}
	v, err := store.INCRBY(k, 1)
	if err != nil {
		resp.WriteError(err)
	} else {
		resp.WriteInt(int(v))
	}
	return resp
}

func (proxy *Proxy) incrby(req *codec.Request) *codec.Response {
	resp := codec.NewResponse()

	if req.ParamsLen() != 2 {
		resp.WriteError(rstore.WrongReqArgsNumber)
		return resp
	}

	k, increment := req.P[0], req.P[1]

	inc, err := strconv.ParseInt(increment, 10, 64)
	if err != nil {
		resp.WriteError(rstore.ParseIntError)
		return resp
	}

	store, err := proxy.doRouter(k)
	if err != nil {
		resp.WriteError(err)
		return resp
	}
	v, err := store.INCRBY(k, inc)
	if err != nil {
		resp.WriteError(err)
	} else {
		resp.WriteInt(int(v))
	}
	return resp
}

func (proxy *Proxy) decr(req *codec.Request) *codec.Response {
	resp := codec.NewResponse()

	if req.ParamsLen() != 1 {
		resp.WriteError(rstore.WrongReqArgsNumber)
		return resp
	}

	k := req.P[0]
	inc := int64(-1)

	store, err := proxy.doRouter(k)
	if err != nil {
		resp.WriteError(err)
		return resp
	}
	v, err := store.INCRBY(k, inc)
	if err != nil {
		resp.WriteError(err)
	} else {
		resp.WriteInt(int(v))
	}
	return resp
}
func (proxy *Proxy) decrby(req *codec.Request) *codec.Response {
	resp := codec.NewResponse()

	if req.ParamsLen() != 2 {
		resp.WriteError(rstore.WrongReqArgsNumber)
		return resp
	}

	k, increment := req.P[0], req.P[1]

	inc, err := strconv.ParseInt(increment, 10, 64)
	if err != nil {
		resp.WriteError(rstore.ParseIntError)
		return resp
	}

	//dec
	inc = -inc

	store, err := proxy.doRouter(k)
	if err != nil {
		resp.WriteError(err)
		return resp
	}
	val, err := store.INCRBY(k, inc)
	if err != nil {
		resp.WriteError(err)
	} else {
		resp.WriteInt(int(val))
	}
	return resp
}

func (proxy *Proxy) hset(req *codec.Request) *codec.Response {
	resp := codec.NewResponse()

	if req.ParamsLen() != 3 {
		resp.WriteError(rstore.WrongReqArgsNumber)
		return resp
	}

	k, hk, v := req.P[0], req.P[1], req.P[2]

	store, err := proxy.doRouter(k)
	if err != nil {
		resp.WriteError(err)
		return resp
	}

	num, err := store.HSET(k, hk, v)
	if err != nil {
		resp.WriteError(err)
	} else {
		resp.WriteInt(num)
	}
	return resp
}

func (proxy *Proxy) hmset(req *codec.Request) *codec.Response {
	resp := codec.NewResponse()

	l := req.ParamsLen()
	if l < 3 || l%2 == 0 {
		resp.WriteError(rstore.WrongReqArgsNumber)
		return resp
	}

	k := req.P[0]
	hkv := make(map[string]string)
	//k,v pairs
	for i := 1; i < l; i += 2 {
		hkv[req.P[i]] = req.P[i+1]
	}

	store, err := proxy.doRouter(k)
	if err != nil {
		resp.WriteError(err)
		return resp
	}

	//the affect number is no used
	_, err = store.HMSET(k, hkv)
	if err != nil {
		resp.WriteError(err)
	} else {
		resp.WriteOK()
	}
	return resp
}

func (proxy *Proxy) hget(req *codec.Request) *codec.Response {
	resp := codec.NewResponse()

	if req.ParamsLen() != 2 {
		resp.WriteError(rstore.WrongReqArgsNumber)
		return resp
	}

	k, hk := req.P[0], req.P[1]

	store, err := proxy.doRouter(k)
	if err != nil {
		resp.WriteError(err)
		return resp
	}

	val, err := store.HGET(k, hk)
	if err != nil {
		if err == rstore.KeyIsNilError {
			resp.WriteNil()
		} else {
			resp.WriteError(err)
		}
	} else {
		resp.WriteString(val)
	}
	return resp
}

func (proxy *Proxy) hgetall(req *codec.Request) *codec.Response {
	resp := codec.NewResponse()

	if req.ParamsLen() != 1 {
		resp.WriteError(rstore.WrongReqArgsNumber)
		return resp
	}

	k := req.P[0]

	store, err := proxy.doRouter(k)
	if err != nil {
		resp.WriteError(err)
		return resp
	}

	hkv, err := store.HGETALL(k)
	if err != nil {
		resp.WriteError(err)
	} else {
		//kv pairs
		out := make([]string, len(hkv)*2)
		var i int = 0
		for k, v := range hkv {
			out[i] = k
			i++
			out[i] = v
			i++
		}
		resp.WriteStringBulk(out)
	}
	return resp
}

func (proxy *Proxy) hmget(req *codec.Request) *codec.Response {
	resp := codec.NewResponse()

	if req.ParamsLen() < 2 {
		resp.WriteError(rstore.WrongReqArgsNumber)
		return resp
	}

	k, hks := req.P[0], req.P[1:]

	store, err := proxy.doRouter(k)
	if err != nil {
		resp.WriteError(err)
		return resp
	}
	hkv, err := store.HMGET(k, hks)
	if err != nil {
		resp.WriteError(err)
	} else {
		//use byte to make nil type
		out := make([][]byte, len(hks)*2)
		for i, k := range hks {
			out[2*i] = []byte(k)
			if v, ok := hkv[k]; ok {
				out[2*i+1] = []byte(v)
			} else {
				out[2*i+1] = nil
			}
		}
		resp.WriteBulk(out)
	}
	return resp
}

func (proxy *Proxy) hdel(req *codec.Request) *codec.Response {
	resp := codec.NewResponse()

	if req.ParamsLen() != 2 {
		resp.WriteError(rstore.WrongReqArgsNumber)
		return resp
	}

	k, hk := req.P[0], req.P[1]
	store, err := proxy.doRouter(k)
	if err != nil {
		resp.WriteError(err)
		return resp
	}

	n, err := store.HDEL(k, hk)
	if err != nil {
		resp.WriteError(err)
	} else {
		resp.WriteInt(n)
	}
	return resp
}

func (proxy *Proxy) hlen(req *codec.Request) *codec.Response {
	resp := codec.NewResponse()

	if req.ParamsLen() != 1 {
		resp.WriteError(rstore.WrongReqArgsNumber)
		return resp
	}

	k := req.P[0]

	store, err := proxy.doRouter(k)
	if err != nil {
		resp.WriteError(err)
		return resp
	}
	n, err := store.HLEN(k)
	if err != nil {
		resp.WriteError(err)
	} else {
		resp.WriteInt(n)
	}
	return resp
}

func (proxy *Proxy) hexists(req *codec.Request) *codec.Response {
	resp := codec.NewResponse()

	if req.ParamsLen() != 2 {
		resp.WriteError(rstore.WrongReqArgsNumber)
		return resp
	}

	k, hk := req.P[0], req.P[1]
	store, err := proxy.doRouter(k)
	if err != nil {
		resp.WriteError(err)
		return resp
	}

	exist, err := store.HEXISTS(k, hk)
	if err != nil {
		resp.WriteError(err)
	} else {
		if exist {
			resp.WriteInt(1)
		} else {

			resp.WriteInt(0)
		}
	}
	return resp
}

func (proxy *Proxy) hkeys(req *codec.Request) *codec.Response {
	resp := codec.NewResponse()

	if req.ParamsLen() != 1 {
		resp.WriteError(rstore.WrongReqArgsNumber)
		return resp
	}

	k := req.P[0]
	store, err := proxy.doRouter(k)
	if err != nil {
		resp.WriteError(err)
		return resp
	}

	ks, err := store.HKEYS(k)
	if err != nil {
		resp.WriteError(err)
	} else {
		resp.WriteStringBulk(ks)
	}
	return resp
}

func (proxy *Proxy) hvals(req *codec.Request) *codec.Response {
	resp := codec.NewResponse()

	if req.ParamsLen() != 1 {
		resp.WriteError(rstore.WrongReqArgsNumber)
		return resp
	}

	k := req.P[0]

	store, err := proxy.doRouter(k)
	if err != nil {
		resp.WriteError(err)
		return resp
	}
	vs, err := store.HVALS(k)
	if err != nil {
		resp.WriteError(err)
	} else {
		resp.WriteStringBulk(vs)
	}
	return resp
}

func (proxy *Proxy) hincrby(req *codec.Request) *codec.Response {
	resp := codec.NewResponse()

	if req.ParamsLen() != 3 {
		resp.WriteError(rstore.WrongReqArgsNumber)
		return resp
	}

	k, hk, increment := req.P[0], req.P[1], req.P[2]

	inc, err := strconv.ParseInt(increment, 10, 64)
	if err != nil {
		resp.WriteError(rstore.ParseIntError)
		return resp
	}

	store, err := proxy.doRouter(k)
	if err != nil {
		resp.WriteError(err)
		return resp
	}
	v, err := store.HINCRBY(k, hk, inc)
	if err != nil {
		resp.WriteError(err)
	} else {
		resp.WriteInt(int(v))
	}
	return resp
}

func (proxy *Proxy) EXISTS(key string) (bool, error) {
	return false, nil
}
func (proxy *Proxy) DEL(key string) error {
	return nil
}

func (proxy *Proxy) EXPIRE(key string, expireSeconds int) (int64, error) {
	return 0, nil
}
func (proxy *Proxy) EXPIREAT(key string, expireAt int) (int64, error) {
	return 0, nil
}
func (proxy *Proxy) TTL(key string) (int64, error) {
	return 0, nil
}
func (proxy *Proxy) TYPE(key string) (string, error) {
	return "", nil
}

func (proxy *Proxy) zadd(req *codec.Request) *codec.Response {
	resp := codec.NewResponse()

	if req.ParamsLen() != 3 {
		resp.WriteError(rstore.WrongReqArgsNumber)
		return resp
	}

	k, s, m := req.P[0], req.P[1], req.P[2]

	score, err := strconv.ParseFloat(s, 64)
	if err != nil {
		resp.WriteError(rstore.ParseFloatError)
		return resp
	}

	store, err := proxy.doRouter(k)
	if err != nil {
		resp.WriteError(err)
		return resp
	}
	n, err := store.ZADD(k, score, m)
	if err != nil {
		resp.WriteError(err)
	} else {
		resp.WriteInt(n)
	}
	return resp
}

func (proxy *Proxy) zscore(req *codec.Request) *codec.Response {
	resp := codec.NewResponse()

	if req.ParamsLen() != 2 {
		resp.WriteError(rstore.WrongReqArgsNumber)
		return resp
	}

	k, m := req.P[0], req.P[1]

	store, err := proxy.doRouter(k)
	if err != nil {
		resp.WriteError(err)
		return resp
	}
	score, err := store.ZSCORE(k, m)
	if err != nil {
		if err == rstore.KeyIsNilError {
			resp.WriteNil()
		} else {
			resp.WriteError(err)
		}
	} else {
		resp.WriteString(score)
	}
	return resp
}

func (proxy *Proxy) zrem(req *codec.Request) *codec.Response {
	resp := codec.NewResponse()

	if req.ParamsLen() != 2 {
		resp.WriteError(rstore.WrongReqArgsNumber)
		return resp
	}

	k, m := req.P[0], req.P[1]
	store, err := proxy.doRouter(k)
	if err != nil {
		resp.WriteError(err)
		return resp
	}

	n, err := store.ZREM(k, m)
	if err != nil {
		resp.WriteError(err)
	} else {
		resp.WriteInt(n)
	}
	return resp
}

func (proxy *Proxy) zremrangebyscore(req *codec.Request) *codec.Response {
	resp := codec.NewResponse()

	if req.ParamsLen() != 3 {
		resp.WriteError(rstore.WrongReqArgsNumber)
		return resp
	}

	k, n1, n2 := req.P[0], req.P[1], req.P[2]

	min, err := strconv.ParseFloat(n1, 64)
	if err != nil {
		resp.WriteError(rstore.ParseFloatError)
		return resp
	}
	max, err := strconv.ParseFloat(n2, 64)
	if err != nil {
		resp.WriteError(rstore.ParseFloatError)
		return resp
	}

	store, err := proxy.doRouter(k)
	if err != nil {
		resp.WriteError(err)
		return resp
	}
	n, err := store.ZREMRANGEBYSCORE(k, min, max)

	if err != nil {
		resp.WriteError(err)
	} else {
		resp.WriteInt(n)
	}
	return resp
}
func (proxy *Proxy) zrange(req *codec.Request) *codec.Response {
	resp := codec.NewResponse()

	l := req.ParamsLen()

	if l != 3 && l != 4 {
		resp.WriteError(rstore.WrongReqArgsNumber)
		return resp
	}

	k, n1, n2 := req.P[0], req.P[1], req.P[2]

	withscores := false
	if l == 4 {
		if strings.ToUpper(req.P[3]) == "WITHSCORES" {
			withscores = true
		} else {
			resp.WriteError(rstore.WrongWithScoresSynax)
			return resp
		}
	}

	min, err := strconv.ParseInt(n1, 10, 64)
	if err != nil {
		resp.WriteError(rstore.ParseIntError)
		return resp
	}
	max, err := strconv.ParseInt(n2, 10, 64)
	if err != nil {
		resp.WriteError(rstore.ParseIntError)
		return resp
	}

	store, err := proxy.doRouter(k)
	if err != nil {
		resp.WriteError(err)
		return resp
	}
	bulks, err := store.ZRANGE(k, int(min), int(max), withscores)

	if err != nil {
		resp.WriteError(err)
	} else {
		resp.WriteStringBulk(bulks)
	}
	return resp
}

func (proxy *Proxy) zrangebyscore(req *codec.Request) *codec.Response {
	resp := codec.NewResponse()

	l := req.ParamsLen()

	if l != 3 && l != 4 {
		resp.WriteError(rstore.WrongReqArgsNumber)
		return resp
	}

	k, n1, n2 := req.P[0], req.P[1], req.P[2]

	withscore := false
	if l == 4 {
		if strings.ToUpper(req.P[3]) == "WITHSCORES" {
			withscore = true
		} else {
			resp.WriteError(rstore.WrongWithScoresSynax)
			return resp
		}
	}

	min, err := strconv.ParseFloat(n1, 64)
	if err != nil {
		resp.WriteError(rstore.ParseFloatError)
		return resp
	}
	max, err := strconv.ParseFloat(n2, 64)
	if err != nil {
		resp.WriteError(rstore.ParseFloatError)
		return resp
	}
	store, err := proxy.doRouter(k)
	if err != nil {
		resp.WriteError(err)
		return resp
	}

	bulk, err := store.ZRANGEBYSCORE(k, min, max, withscore)

	if err != nil {
		resp.WriteError(err)
	} else {
		resp.WriteStringBulk(bulk)
	}
	return resp
}

func (proxy *Proxy) zcard(req *codec.Request) *codec.Response {
	resp := codec.NewResponse()

	if req.ParamsLen() != 1 {
		resp.WriteError(rstore.WrongReqArgsNumber)
		return resp
	}

	k := req.P[0]

	store, err := proxy.doRouter(k)
	if err != nil {
		resp.WriteError(err)
		return resp
	}
	n, err := store.ZCARD(k)
	if err != nil {
		resp.WriteError(err)
	} else {
		resp.WriteInt(n)
	}
	return resp
}

func (proxy *Proxy) zcount(req *codec.Request) *codec.Response {
	resp := codec.NewResponse()
	if req.ParamsLen() != 3 {
		resp.WriteError(rstore.WrongReqArgsNumber)
		return resp
	}

	k, n1, n2 := req.P[0], req.P[1], req.P[2]

	min, err := strconv.ParseFloat(n1, 64)
	if err != nil {
		resp.WriteError(rstore.ParseFloatError)
		return resp
	}
	max, err := strconv.ParseFloat(n2, 64)
	if err != nil {
		resp.WriteError(rstore.ParseFloatError)
		return resp
	}

	store, err := proxy.doRouter(k)
	if err != nil {
		resp.WriteError(err)
		return resp
	}
	n, err := store.ZCOUNT(k, min, max)
	if err != nil {
		resp.WriteError(err)
	} else {
		resp.WriteInt(n)
	}
	return resp
}

func (proxy *Proxy) zrank(req *codec.Request) *codec.Response {
	resp := codec.NewResponse()
	if req.ParamsLen() != 2 {
		resp.WriteError(rstore.WrongReqArgsNumber)
		return resp
	}

	k, m := req.P[0], req.P[1]

	store, err := proxy.doRouter(k)
	if err != nil {
		resp.WriteError(err)
		return resp
	}
	n, err := store.ZRANK(k, m)
	if err != nil {
		resp.WriteError(err)
	} else {
		if n > 0 {
			resp.WriteInt(n)
		} else {
			resp.WriteNil()
		}
	}
	return resp
}

func (proxy *Proxy) proxyZrevRangeWithScore(req *codec.Request) *codec.Response {
	resp := codec.NewResponse()
	return resp
}

func (proxy *Proxy) proxySadd(req *codec.Request) *codec.Response {
	resp := codec.NewResponse()
	return resp
}
func (proxy *Proxy) proxyScard(req *codec.Request) *codec.Response {
	resp := codec.NewResponse()
	return resp
}
func (proxy *Proxy) proxySisMember(req *codec.Request) *codec.Response {
	resp := codec.NewResponse()
	return resp
}

func (proxy *Proxy) proxySmembers(req *codec.Request) *codec.Response {
	resp := codec.NewResponse()
	return resp
}
func (proxy *Proxy) proxySrem(req *codec.Request) *codec.Response {
	resp := codec.NewResponse()
	return resp
}