package util_test

import (
	"testing"

	"github.com/nopryskua/event-catalogue/backend/internal/util"
	"github.com/stretchr/testify/require"
)

type MyType bool

type MyTypeNamer bool

func (MyTypeNamer) TypeName() string {
	return "Override"
}

func TestTypeName(t *testing.T) {
	require.Equal(t, "github.com/nopryskua/event-catalogue/backend/internal/util_test.MyType", util.TypeName[MyType]())
	require.Equal(t, "Override", util.TypeName[MyTypeNamer]())
}
