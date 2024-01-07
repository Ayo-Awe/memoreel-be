package datastore

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindRecipient(t *testing.T) {
	reel := Reel{Recipients: []Recipient{{UID: "12345"}, {UID: "123456"}, {UID: "12346"}}}

	r1ID := "12345"
	r1 := reel.FindRecipient(r1ID)
	require.NotNil(t, r1)
	require.Equal(t, r1.UID, r1ID)

	r2ID := "123489"
	r2 := reel.FindRecipient(r2ID)
	require.Nil(t, r2)

}
