package persistence

import (
    "net"
    "testing"
)

func TestNextIP(t *testing.T) {
    ip := net.ParseIP("192.168.0.1").To4()
    next := nextIP(ip)
    want := net.ParseIP("192.168.0.2").To4()
    if !next.Equal(want) {
        t.Errorf("nextIP(192.168.0.1) = %v; want %v", next, want)
    }
    // test rollover
    ip = net.ParseIP("192.168.0.255").To4()
    next = nextIP(ip)
    want = net.ParseIP("192.168.1.0").To4()
    if !next.Equal(want) {
        t.Errorf("nextIP(192.168.0.255) = %v; want %v", next, want)
    }
}
