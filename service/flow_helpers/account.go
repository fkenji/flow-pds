package flow_helpers

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/crypto"
)

var accounts map[flow.Address]*Account
var accountsLock = &sync.Mutex{} // Making sure our "accounts" var is a singleton
var keyIndexLock = &sync.Mutex{}

var seqNumLock = &sync.Mutex{}
var seqNumMap map[flow.Address]map[int]uint64

type Account struct {
	Address           flow.Address
	PrivateKeyInHex   string
	KeyIndexes        []int
	nextKeyIndexIndex int
}

// GetAccount either returns an Account from the application wide cache or initiliazes a new Account
func GetAccount(address flow.Address, privateKeyInHex string, keyIndexes []int) *Account {
	accountsLock.Lock()
	defer accountsLock.Unlock()

	if accounts == nil {
		accounts = make(map[flow.Address]*Account, 1)
	}

	if existing, ok := accounts[address]; ok {
		return existing
	}

	// Pick a random index to start from
	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(keyIndexes))

	new := &Account{
		Address:           address,
		PrivateKeyInHex:   privateKeyInHex,
		KeyIndexes:        keyIndexes,
		nextKeyIndexIndex: randomIndex,
	}

	accounts[address] = new

	return new
}

// KeyIndex rotates the given indexes ('KeyIndexes') and returns the next index
// TODO (latenssi): sync over database as this currently only works in a single instance situation
func (a *Account) KeyIndex() int {
	// NOTE: This won't help if having multiple instances of the PDS service running
	keyIndexLock.Lock()
	defer keyIndexLock.Unlock()

	i := a.KeyIndexes[a.nextKeyIndexIndex]
	a.nextKeyIndexIndex = (a.nextKeyIndexIndex + 1) % len(a.KeyIndexes)

	return i
}

func (a Account) GetProposalKey(ctx context.Context, flowClient *client.Client) (*flow.AccountKey, error) {
	account, err := flowClient.GetAccount(ctx, a.Address)
	k := account.Keys[a.KeyIndex()]
	if err != nil {
		return nil, fmt.Errorf("error in flow_helpers.Account.GetProposalKey: %w", err)
	}
	k.SequenceNumber = getSeqNum(a.Address, k)
	return k, nil
}

func (a Account) GetSigner() (crypto.Signer, error) {
	p, err := crypto.DecodePrivateKeyHex(crypto.ECDSA_P256, a.PrivateKeyInHex)
	if err != nil {
		return nil, fmt.Errorf("error in flow_helpers.Account.GetSigner: %w", err)
	}
	return crypto.NewNaiveSigner(p, crypto.SHA3_256), nil
}

// getSeqNum, is a hack around the fact that GetAccount on Flow Client returns
// the latest SequenceNumber on-chain but it might be outdated as we may be
// sending multiple transactions in the current block
// TODO (latenssi): sync over database as this currently only works in a single instance situation
func getSeqNum(address flow.Address, key *flow.AccountKey) uint64 {
	seqNumLock.Lock()
	defer seqNumLock.Unlock()

	if seqNumMap == nil {
		seqNumMap = make(map[flow.Address]map[int]uint64)
	}

	if seqNumMap[address] == nil {
		seqNumMap[address] = make(map[int]uint64)
	}

	if prev, ok := seqNumMap[address][key.Index]; ok && prev >= key.SequenceNumber {
		seqNumMap[address][key.Index]++
	} else {
		seqNumMap[address][key.Index] = key.SequenceNumber
	}

	return seqNumMap[address][key.Index]
}