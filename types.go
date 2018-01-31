package rrgo

import (
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/sha3"
)

type Address [20]byte

type Uint256 [32]byte

func HexStringToBytes(hexString string) ([]byte, error) {
	return hex.DecodeString(strings.TrimPrefix(hexString, "0x"))
}

func IntStringToBytes(intString string) ([]byte, error) {
	bigInt := new(big.Int)
	_, success := bigInt.SetString(intString, 10)
	if success {
		return abi.U256(bigInt), nil
	}
	return nil, errors.New("Value not a valid integer")

}

func (addr *Address) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		copy(addr[:], v)
		return nil
	default:
		return errors.New("Address scanner src should be []byte")
	}
}

func (addr *Uint256) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		copy(addr[:], v)
		return nil
	default:
		return errors.New("Uint256 scanner src should be []byte")
	}
}

type TokenAddress struct {
	Address string `json:"address"`
	Symbol  string `json:"symbol"`
}

func (u *Uint256) UnmarshalJSON(b []byte) error {
	bs, err := IntStringToBytes(strings.Trim(string(b), `"`))
	if err != nil {
		return err
	}
	copy(u[:], bs)
	return nil
}

/*
func (a *Address) MarshalJSON() ([]byte, error) {
	return nil, nil

}
func (a *Address) UnmarshalJSON(b []byte) error {
	bs, err := HexStringToBytes(strings.Trim(string(b), `"`))
	if err != nil {
		return err
	}
	copy(a[:], bs)
	return nil
}
*/

type jsonToken struct {
	Address   string `json:"address"`
	MinAmount string `json:"minAmount"`
	MaxAmount string `json:"maxAmount"`
	Precision uint64 `json:"precision"`
}

type Token struct {
	Address   string
	MinAmount *Uint256
	MaxAmount *Uint256
	Precision uint64
}

type PairsOpts struct {
	TokenA string `url:"tokenA,omitempty"`
	TokenB string `url:"tokenB,omitempty"`
}

type OrdersOpts struct {
	ExchangeAddress string `url:"exchangeContractAddress,omitempty"`
	TokenAddress    string `url:"tokenAddress,omitempty"`
	MakerToken      string `url:"makerTokenAddress,omitempty"`
	TakerToken      string `url:"takerTokenAddress,omitempty"`
	PairsOpts
	Maker        string `url:"maker,omitempty"`
	Taker        string `url:"taker,omitempty"`
	Trader       string `url:"trader,omitempty"`
	FeeRecipient string `url:"feeRecipient,omitempty"`
}

type OrderbookOpts struct {
	BaseTokenAddress  string `url:"baseTokenAddress"`
	QuoteTokenAddress string `url:"quoteTokenAddress"`
}

type Orderbook struct {
	Bids []APIOrder `json:"bids"`
	Asks []APIOrder `json:"asks"`
}

type jsonPair struct {
	TokenA jsonToken `json:"tokenA,omitempty"`
	TokenB jsonToken `json:"tokenB,omitempty"`
}

type Pair struct {
	TokenA *Token `json:"tokenA,omitempty"`
	TokenB *Token `json:"tokenB,omitempty"`
}

type Response struct {
	*http.Response
	Rate
}

type Rate struct {
	RequestLimit      int       `json:"request_limit"`
	RequestsRemaining int       `json:"requests_remaining"`
	Reset             Timestamp `json:"rate_reset"`
}

// Timestamp represents a time that can be unmarshalled from a JSON string
// formatted as either an RFC3339 or Unix timestamp. All
// exported methods of time.Time can be called on Timestamp.
type Timestamp struct {
	time.Time
}

func (t Timestamp) String() string {
	return t.Time.String()
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// Time is expected in RFC3339 or Unix format.
func (t *Timestamp) UnmarshalJSON(data []byte) (err error) {
	str := string(data)
	i, err := strconv.ParseInt(str, 10, 64)
	if err == nil {
		t.Time = time.Unix(i, 0)
	} else {
		t.Time, err = time.Parse(`"`+time.RFC3339+`"`, str)
	}
	return
}

// Equal reports whether t and u are equal based on time.Equal
func (t Timestamp) Equal(u Timestamp) bool {
	return t.Time.Equal(u.Time)
}

// Order represents an 0x order object
type Order struct {
	Maker                     *Address
	Taker                     *Address
	MakerToken                *Address
	TakerToken                *Address
	FeeRecipient              *Address
	ExchangeAddress           *Address
	MakerTokenAmount          *Uint256
	TakerTokenAmount          *Uint256
	MakerFee                  *Uint256
	TakerFee                  *Uint256
	ExpirationTimestampInSec  *Uint256
	Salt                      *Uint256
	Signature                 *Signature
	TakerTokenAmountFilled    *Uint256
	TakerTokenAmountCancelled *Uint256
}

func (order *Order) Initialize() {
	order.ExchangeAddress = &Address{}
	order.Maker = &Address{}
	order.Taker = &Address{}
	order.MakerToken = &Address{}
	order.TakerToken = &Address{}
	order.FeeRecipient = &Address{}
	order.MakerTokenAmount = &Uint256{}
	order.TakerTokenAmount = &Uint256{}
	order.MakerFee = &Uint256{}
	order.TakerFee = &Uint256{}
	order.ExpirationTimestampInSec = &Uint256{}
	order.Salt = &Uint256{}
	order.TakerTokenAmountFilled = &Uint256{}
	order.TakerTokenAmountCancelled = &Uint256{}
	order.Signature = &Signature{}
}

// NewOrder takes string representations of values and converts them into an Order object
func NewOrder(maker, taker, makerToken, takerToken, feeRecipient, exchangeAddress, makerTokenAmount, takerTokenAmount, makerFee, takerFee, expirationTimestampInSec, salt, sigV, sigR, sigS, takerTokenAmountFilled, takerTokenAmountCancelled string) (*Order, error) {
	order := Order{}
	if err := order.fromStrings(maker, taker, makerToken, takerToken, feeRecipient, exchangeAddress, makerTokenAmount, takerTokenAmount, makerFee, takerFee, expirationTimestampInSec, salt, sigV, sigR, sigS, takerTokenAmountFilled, takerTokenAmountCancelled); err != nil {
		return nil, err
	}
	return &order, nil
}

func (order *Order) fromStrings(maker, taker, makerToken, takerToken, feeRecipient, exchangeAddress, makerTokenAmount, takerTokenAmount, makerFee, takerFee, expirationTimestampInSec, salt, sigV, sigR, sigS, takerTokenAmountFilled, takerTokenAmountCancelled string) error {
	order.Initialize()
	makerBytes, err := HexStringToBytes(maker)
	if err != nil {
		return err
	}
	takerBytes, err := HexStringToBytes(taker)
	if err != nil {
		return err
	}
	makerTokenBytes, err := HexStringToBytes(makerToken)
	if err != nil {
		return err
	}
	takerTokenBytes, err := HexStringToBytes(takerToken)
	if err != nil {
		return err
	}
	feeRecipientBytes, err := HexStringToBytes(feeRecipient)
	if err != nil {
		return err
	}
	exchangeAddressBytes, err := HexStringToBytes(exchangeAddress)
	if err != nil {
		return err
	}
	makerTokenAmountBytes, err := IntStringToBytes(makerTokenAmount)
	if err != nil {
		return err
	}
	takerTokenAmountBytes, err := IntStringToBytes(takerTokenAmount)
	if err != nil {
		return err
	}
	makerFeeBytes, err := IntStringToBytes(makerFee)
	if err != nil {
		return err
	}
	takerFeeBytes, err := IntStringToBytes(takerFee)
	if err != nil {
		return err
	}
	expirationTimestampInSecBytes, err := IntStringToBytes(expirationTimestampInSec)
	if err != nil {
		return err
	}
	saltBytes, err := IntStringToBytes(salt)
	if err != nil {
		return err
	}
	sigVInt, err := strconv.Atoi(sigV)
	if err != nil {
		return err
	}
	sigRBytes, err := HexStringToBytes(sigR)
	if err != nil {
		return err
	}
	sigSBytes, err := HexStringToBytes(sigS)
	if err != nil {
		return err
	}
	takerTokenAmountFilledBytes, err := IntStringToBytes(takerTokenAmountFilled)
	if err != nil {
		return err
	}
	takerTokenAmountCancelledBytes, err := IntStringToBytes(takerTokenAmountCancelled)
	if err != nil {
		return err
	}
	copy(order.Maker[:], makerBytes)
	copy(order.Taker[:], takerBytes)
	copy(order.MakerToken[:], makerTokenBytes)
	copy(order.TakerToken[:], takerTokenBytes)
	copy(order.FeeRecipient[:], feeRecipientBytes)
	copy(order.ExchangeAddress[:], exchangeAddressBytes)
	copy(order.MakerTokenAmount[:], makerTokenAmountBytes)
	copy(order.TakerTokenAmount[:], takerTokenAmountBytes)
	copy(order.MakerFee[:], makerFeeBytes)
	copy(order.TakerFee[:], takerFeeBytes)
	copy(order.ExpirationTimestampInSec[:], expirationTimestampInSecBytes)
	copy(order.Salt[:], saltBytes)
	order.Signature = &Signature{}
	order.Signature.V = byte(sigVInt)
	copy(order.Signature.S[:], sigSBytes)
	copy(order.Signature.R[:], sigRBytes)
	copy(order.Signature.Hash[:], order.Hash())
	copy(order.TakerTokenAmountFilled[:], takerTokenAmountFilledBytes)
	copy(order.TakerTokenAmountCancelled[:], takerTokenAmountCancelledBytes)
	return nil
}

func (order *Order) Hash() []byte {
	sha := sha3.NewKeccak256()

	sha.Write(order.ExchangeAddress[:])
	sha.Write(order.Maker[:])
	sha.Write(order.Taker[:])
	sha.Write(order.MakerToken[:])
	sha.Write(order.TakerToken[:])
	sha.Write(order.FeeRecipient[:])
	sha.Write(order.MakerTokenAmount[:])
	sha.Write(order.TakerTokenAmount[:])
	sha.Write(order.MakerFee[:])
	sha.Write(order.TakerFee[:])
	sha.Write(order.ExpirationTimestampInSec[:])
	sha.Write(order.Salt[:])
	return sha.Sum(nil)
}

type APIOrder struct {
	Maker                     string       `json:"maker"`
	Taker                     string       `json:"taker"`
	MakerToken                string       `json:"makerTokenAddress"`
	TakerToken                string       `json:"takerTokenAddress"`
	FeeRecipient              string       `json:"feeRecipient"`
	ExchangeAddress           string       `json:"exchangeContractAddress"`
	MakerTokenAmount          string       `json:"makerTokenAmount"`
	TakerTokenAmount          string       `json:"takerTokenAmount"`
	MakerFee                  string       `json:"makerFee"`
	TakerFee                  string       `json:"takerFee"`
	ExpirationTimestampInSec  string       `json:"expirationUnixTimestampSec"`
	Salt                      string       `json:"salt"`
	Signature                 APISignature `json:"ecSignature"`
	TakerTokenAmountFilled    string       `json:"-"`
	TakerTokenAmountCancelled string       `json:"-"`
}

func (order *Order) UnmarshalJSON(b []byte) error {
	jOrder := APIOrder{}
	if err := json.Unmarshal(b, &jOrder); err != nil {
		return err
	}
	if jOrder.TakerTokenAmountFilled == "" {
		jOrder.TakerTokenAmountFilled = "0"
	}
	if jOrder.TakerTokenAmountCancelled == "" {
		jOrder.TakerTokenAmountCancelled = "0"
	}
	order.fromStrings(
		jOrder.Maker,
		jOrder.Taker,
		jOrder.MakerToken,
		jOrder.TakerToken,
		jOrder.FeeRecipient,
		jOrder.ExchangeAddress,
		jOrder.MakerTokenAmount,
		jOrder.TakerTokenAmount,
		jOrder.MakerFee,
		jOrder.TakerFee,
		jOrder.ExpirationTimestampInSec,
		jOrder.Salt,
		string(jOrder.Signature.V),
		jOrder.Signature.R,
		jOrder.Signature.S,
		jOrder.TakerTokenAmountFilled,
		jOrder.TakerTokenAmountCancelled,
	)

	return nil
}

func (order *Order) MarshalJSON() ([]byte, error) {
	APIOrder := &APIOrder{}
	APIOrder.Maker = fmt.Sprintf("%#x", order.Maker[:])
	APIOrder.Taker = fmt.Sprintf("%#x", order.Taker[:])
	APIOrder.MakerToken = fmt.Sprintf("%#x", order.MakerToken[:])
	APIOrder.TakerToken = fmt.Sprintf("%#x", order.TakerToken[:])
	APIOrder.FeeRecipient = fmt.Sprintf("%#x", order.FeeRecipient[:])
	APIOrder.ExchangeAddress = fmt.Sprintf("%#x", order.ExchangeAddress[:])
	APIOrder.MakerTokenAmount = new(big.Int).SetBytes(order.MakerTokenAmount[:]).String()
	APIOrder.TakerTokenAmount = new(big.Int).SetBytes(order.TakerTokenAmount[:]).String()
	APIOrder.MakerFee = new(big.Int).SetBytes(order.MakerFee[:]).String()
	APIOrder.TakerFee = new(big.Int).SetBytes(order.TakerFee[:]).String()
	APIOrder.ExpirationTimestampInSec = new(big.Int).SetBytes(order.ExpirationTimestampInSec[:]).String()
	APIOrder.Salt = new(big.Int).SetBytes(order.Salt[:]).String()
	APIOrder.Signature = APISignature{}
	APIOrder.Signature.R = fmt.Sprintf("%#x", order.Signature.R[:])
	APIOrder.Signature.V = json.Number(fmt.Sprintf("%v", order.Signature.V))
	APIOrder.Signature.S = fmt.Sprintf("%#x", order.Signature.S[:])
	APIOrder.TakerTokenAmountFilled = new(big.Int).SetBytes(order.TakerTokenAmountFilled[:]).String()
	APIOrder.TakerTokenAmountCancelled = new(big.Int).SetBytes(order.TakerTokenAmountCancelled[:]).String()
	return json.Marshal(APIOrder)
}

func (order *Order) Bytes() [441]byte {
	var output [441]byte
	copy(output[0:20], order.ExchangeAddress[:])             // 20
	copy(output[20:40], order.Maker[:])                      // 20
	copy(output[40:60], order.Taker[:])                      // 20
	copy(output[60:80], order.MakerToken[:])                 // 20
	copy(output[80:100], order.TakerToken[:])                // 20
	copy(output[100:120], order.FeeRecipient[:])             // 20
	copy(output[120:152], order.MakerTokenAmount[:])         // 32
	copy(output[152:184], order.TakerTokenAmount[:])         // 32
	copy(output[184:216], order.MakerFee[:])                 // 32
	copy(output[216:248], order.TakerFee[:])                 // 32
	copy(output[248:280], order.ExpirationTimestampInSec[:]) // 32
	copy(output[280:312], order.Salt[:])                     // 32
	output[312] = order.Signature.V
	copy(output[313:345], order.Signature.R[:])
	copy(output[345:377], order.Signature.S[:])
	copy(output[377:409], order.TakerTokenAmountFilled[:])
	copy(output[409:441], order.TakerTokenAmountCancelled[:])
	return output
}

func (o APIOrder) String() string {
	mat, ok := A2T[o.MakerToken]
	if !ok {
		log.Println(o.MakerToken, "is unknown token")
	}
	tat, ok := A2T[o.TakerToken]
	if !ok {
		log.Println(o.TakerToken, "is unknown token")
	}
	mto := o.MakerTokenAmount
	tto := o.TakerTokenAmount

	return fmt.Sprintf("%s %s -> %s %s", mto, mat, tto, tat)
}

func (order *Order) FromBytes(data [441]byte) {
	order.Initialize()
	copy(order.ExchangeAddress[:], data[0:20])
	copy(order.Maker[:], data[20:40])
	copy(order.Taker[:], data[40:60])
	copy(order.MakerToken[:], data[60:80])
	copy(order.TakerToken[:], data[80:100])
	copy(order.FeeRecipient[:], data[100:120])
	copy(order.MakerTokenAmount[:], data[120:152])
	copy(order.TakerTokenAmount[:], data[152:184])
	copy(order.MakerFee[:], data[184:216])
	copy(order.TakerFee[:], data[216:248])
	copy(order.ExpirationTimestampInSec[:], data[248:280])
	copy(order.Salt[:], data[280:312])
	order.Signature = &Signature{}
	order.Signature.V = data[312]
	copy(order.Signature.R[:], data[313:345])
	copy(order.Signature.S[:], data[345:377])
	copy(order.Signature.Hash[:], order.Hash())
	copy(order.TakerTokenAmountFilled[:], data[377:409])
	copy(order.TakerTokenAmountCancelled[:], data[409:441])
}

func OrderFromBytes(data [441]byte) *Order {
	order := Order{}
	order.FromBytes(data)
	return &order
}

type Signature struct {
	V    byte
	R    [32]byte
	S    [32]byte
	Hash [32]byte
}

type APISignature struct {
	V    json.Number `json:"v"`
	R    string      `json:"r"`
	S    string      `json:"s"`
	Hash string      `json:"-"`
}

func (sig *Signature) Verify(address *Address) bool {
	sigValue, _ := sig.Value()
	sigBytes := sigValue.([]byte)

	hashedBytes := append([]byte("\x19Ethereum Signed Message:\n32"), sig.Hash[:]...)
	signedBytes := crypto.Keccak256(hashedBytes)
	pub, err := crypto.Ecrecover(signedBytes, sigBytes)
	if err != nil {
		log.Println(err.Error())
		return false
	}
	recoverAddress := common.BytesToAddress(crypto.Keccak256(pub[1:])[12:])
	return reflect.DeepEqual(address[:], recoverAddress[:])
}

func (sig *Signature) Value() (driver.Value, error) {
	sigBytes := make([]byte, 65)
	copy(sigBytes[32-len(sig.R):32], sig.R[:])
	copy(sigBytes[64-len(sig.S):64], sig.S[:])
	sigBytes[64] = byte(int(sig.V) - 27)
	return sigBytes, nil
}

func (sig *Signature) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		if len(v) != 65 {
			return errors.New("Signature scanner src should be []byte of length 65")
		}
		copy(sig.R[:], v[0:32])
		copy(sig.S[:], v[32:64])
		sig.V = byte(int(v[64]) + 27)
		return nil
	default:
		return errors.New("Signature scanner src should be []byte of length 65")
	}
}
