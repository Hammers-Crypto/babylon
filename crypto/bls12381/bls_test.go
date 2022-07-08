package bls12381

import (
	"crypto/rand"
	"github.com/stretchr/testify/require"
	"testing"

	blst "github.com/supranational/blst/bindings/go"
)

// Tests single BLS sig verification
func TestVerifyBlsSig(t *testing.T) {
	msga := []byte("aaaaaaaa")
	msgb := []byte("bbbbbbbb")
	sk, pk := genRandomKeyPair()
	sig := Sign(sk, msga)
	// a byte size of a sig (compressed) is 48
	require.Equal(t, 48, len(sig))
	// a byte size of a public key (compressed) is 96
	require.Equal(t, 96, len(pk))
	res, err := Verify(sig, pk, msga)
	require.True(t, res)
	require.Nil(t, err)
	res, err = Verify(sig, pk, msgb)
	require.False(t, res)
	require.Nil(t, err)
}

// Tests BLS multi sig verification
func TestVerifyBlsMultiSig(t *testing.T) {
	msga := []byte("aaaaaaaa")
	msgb := []byte("bbbbbbbb")
	n := 100
	sks, pks := generateBatchTestKeyPairs(n)
	sigs := make([]Signature, n)
	for i := 0; i < n; i++ {
		sigs[i] = Sign(sks[i], msga)
	}
	multiSig, err := AggrSigList(sigs)
	require.Nil(t, err)
	res, err := VerifyMultiSig(multiSig, pks, msga)
	require.True(t, res)
	require.Nil(t, err)
	res, err = VerifyMultiSig(multiSig, pks, msgb)
	require.False(t, res)
	require.Nil(t, err)
}

// Tests BLS multi sig verification
// insert an invalid BLS sig in aggregation
func TestVerifyBlsMultiSig2(t *testing.T) {
	msga := []byte("aaaaaaaa")
	msgb := []byte("bbbbbbbb")
	n := 100
	sks, pks := generateBatchTestKeyPairs(n)
	sigs := make([]Signature, n)
	for i := 0; i < n-1; i++ {
		sigs[i] = Sign(sks[i], msga)
	}
	sigs[n-1] = Sign(sks[n-1], msgb)
	multiSig, err := AggrSigList(sigs)
	require.Nil(t, err)
	res, err := VerifyMultiSig(multiSig, pks, msga)
	require.False(t, res)
	require.Nil(t, err)
	res, err = VerifyMultiSig(multiSig, pks, msgb)
	require.False(t, res)
	require.Nil(t, err)
}

func TestAccumulativeAggregation(t *testing.T) {
	msga := []byte("aaaaaaaa")
	msgb := []byte("bbbbbbbb")
	n := 100
	sks, pks := generateBatchTestKeyPairs(n)
	var aggPK PublicKey
	var aggSig Signature
	var err error
	var res bool
	for i := 0; i < n-1; i++ {
		sig := Sign(sks[i], msga)
		aggSig, err = AggrSig(aggSig, sig)
		require.Nil(t, err)
		aggPK, err = AggrPK(aggPK, pks[i])
		require.Nil(t, err)
		res, err = Verify(aggSig, aggPK, msga)
		require.True(t, res)
		require.Nil(t, err)
	}
	sig := Sign(sks[n-1], msgb)
	aggSig, err = AggrSig(aggSig, sig)
	aggPK, err = AggrPK(aggPK, pks[n-1])
	res, err = Verify(aggSig, aggPK, msga)
	require.False(t, res)
	require.Nil(t, err)
}

func genRandomKeyPair() (*blst.SecretKey, PublicKey) {
	var ikm [32]byte
	_, _ = rand.Read(ikm[:])
	return GenKeyPair(ikm[:])
}

func generateBatchTestKeyPairs(n int) ([]*blst.SecretKey, []PublicKey) {
	sks := make([]*blst.SecretKey, n)
	pubks := make([]PublicKey, n)
	for i := 0; i < n; i++ {
		sk, pk := genRandomKeyPair()
		sks[i] = sk
		pubks[i] = pk
	}
	return sks, pubks
}
