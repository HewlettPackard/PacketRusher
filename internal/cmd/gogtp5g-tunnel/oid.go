package tunnel

import (
	"strconv"
	"strings"

	"github.com/free5gc/go-gtp5gnl"
)

// oid <- (seid ':' id) / id
func ParseOID(s string) (gtp5gnl.OID, error) {
	i := strings.IndexRune(s, ':')
	if i == -1 {
		id, err := strconv.ParseUint(s, 10, 32)
		if err != nil {
			return nil, err
		}
		return gtp5gnl.OID{id}, nil
	}
	seid, err := strconv.ParseUint(s[:i], 10, 64)
	if err != nil {
		return nil, err
	}
	id, err := strconv.ParseUint(s[i+1:], 10, 32)
	if err != nil {
		return nil, err
	}
	return gtp5gnl.OID{seid, id}, nil
}
