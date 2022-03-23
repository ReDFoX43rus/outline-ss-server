package features

import (
	"container/list"
	"fmt"
	"net"
	"sync"
)

type ConnLimit interface {
	CanEstablishNewConnection(id string, clientIP net.IP) bool

	OnConnectionEstablished(id string, clientIP net.IP) bool

	OnConnectionLost(id string, clientIP net.IP)
}

type AllowedConnections struct {
	Id             string
	MaxConnections uint32
}

type connectedIP struct {
	ip    net.IP
	count uint32
}

type connLimitMap struct {
	connectedIPs          map[string](*list.List)
	maxAllowedConnections map[string]uint32

	mutex    sync.RWMutex
	debugLog bool
}

func MakeConnLimitInfo(allowedConnections []AllowedConnections, enableDebugLog bool) ConnLimit {
	var connectionsMap = make(map[string](*list.List))
	var maxConnectionsMap = make(map[string]uint32)

	for _, e := range allowedConnections {
		maxConnectionsMap[e.Id] = e.MaxConnections
	}

	return &connLimitMap{
		connectedIPs:          connectionsMap,
		maxAllowedConnections: maxConnectionsMap,
		debugLog:              enableDebugLog,
	}
}

func (c *connLimitMap) CanEstablishNewConnection(id string, clientIP net.IP) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	size := c.connectedIPsForID(id)
	hasConnection := c.alreadyHasConnection(id, clientIP)
	result := hasConnection || size < int(c.maxAllowedConnections[id])

	if c.debugLog {
		fmt.Printf("CanEstablishNewConnection, id=%s, ip=%s, can=%t, hasConnection=%t, size=%d, max=%d\n", id, clientIP, result, hasConnection, size, c.maxAllowedConnections[id])
	}

	return result
}

func (c *connLimitMap) OnConnectionEstablished(id string, clientIP net.IP) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	nowConnections := c.incrementConnectionsCounter(id, clientIP)

	if c.debugLog {
		fmt.Printf("OnConnectionEstablished, id=%s, ip=%s, nowConnections=%d\n", id, clientIP, nowConnections)
	}

	return true
}

func (c *connLimitMap) OnConnectionLost(id string, clientIP net.IP) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	nowConnections := c.decrementConnectionsCounter(id, clientIP)

	if c.debugLog {
		fmt.Printf("OnConnectionLost, id=%s, ip=%s, nowConnections=%d\n", id, clientIP, nowConnections)
	}
}

func (c *connLimitMap) alreadyHasConnection(id string, clientIP net.IP) bool {
	list := c.connectedIPs[id]
	if list != nil {
		for e := list.Front(); e != nil; e = e.Next() {
			if e.Value != nil {
				cip := e.Value.(*connectedIP)
				if cip.ip.Equal(clientIP) {
					return cip.count > 0
				}
			}
		}
	}

	return false
}

func (c *connLimitMap) connectedIPsForID(id string) int {
	connectedIPs := c.connectedIPs[id]

	var size int
	if connectedIPs == nil {
		size = 0
	} else {
		for e := connectedIPs.Front(); e != nil; e = e.Next() {
			if e.Value != nil {
				count := e.Value.(*connectedIP).count
				if count > 0 {
					size += 1
				}
			}
		}
	}

	return size
}

func (c *connLimitMap) incrementConnectionsCounter(id string, ip net.IP) uint32 {
	var connectedIPs *list.List

	if c.connectedIPs[id] != nil {
		connectedIPs = c.connectedIPs[id]
	} else {
		connectedIPs = list.New()
		c.connectedIPs[id] = connectedIPs
	}

	var connectedIPStruct *connectedIP

	for e := connectedIPs.Front(); e != nil; e = e.Next() {
		if e.Value.(*connectedIP).ip.Equal(ip) {
			connectedIPStruct = e.Value.(*connectedIP)
			break
		}
	}

	if connectedIPStruct == nil {
		connectedIPStruct = &connectedIP{
			ip:    ip,
			count: 0,
		}

		connectedIPs.PushBack(connectedIPStruct)
	}

	connectedIPStruct.count += 1
	return connectedIPStruct.count
}

func (c *connLimitMap) decrementConnectionsCounter(id string, ip net.IP) uint32 {
	connectedIPs := c.connectedIPs[id]

	if connectedIPs == nil {
		panic(fmt.Sprintf("Trying to decrement unexisting id value, id=%s, ip=%s", id, ip))
	}

	var connectedIPStruct *connectedIP

	for e := connectedIPs.Front(); e != nil; e = e.Next() {
		if e.Value.(*connectedIP).ip.Equal(ip) {
			connectedIPStruct = e.Value.(*connectedIP)
			break
		}
	}

	if connectedIPStruct == nil {
		panic(fmt.Sprintf("Trying to decrement unconnected ip value, id=%s, ip=%s", id, ip))
	}

	if connectedIPStruct.count <= 0 {
		panic(fmt.Sprintf("Trying to decrement 0 counter, id=%s, ip=%s", id, ip))
	}

	connectedIPStruct.count -= 1
	return connectedIPStruct.count
}
