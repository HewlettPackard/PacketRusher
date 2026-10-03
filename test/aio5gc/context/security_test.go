package context

import (
	nas "github.com/free5gc/nas/message"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSecurityCountersAreIndependent(t *testing.T) {
	var s SecurityContext
	s.SetULCount(nas.Count{Count: 42})
	s.SetDLCount(nas.Count{Count: 99})
	require.Equal(t, uint32(42), s.GetULCount().Count)
	require.Equal(t, uint32(99), s.GetDLCount().Count)
}
