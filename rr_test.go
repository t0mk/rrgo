package rrgo

import (
	"log"
	"os"
	"strings"
	"testing"
)

func TestTokenPairs(t *testing.T) {
	c := NewClient()
	pr := PairsOpts{TokenA: "WETH"}
	pairs, _, err := c.Pairs(pr)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("WETH", len(pairs))

	pairs, _, err = c.Pairs(PairsOpts{})
	if err != nil {
		t.Fatal(err)
	}
	log.Println("all", len(pairs))
	for _, p := range pairs {
		log.Printf("%s/%s\n",
			A2T[p.TokenA.Address],
			A2T[p.TokenB.Address])
	}
}

func TestOrders(t *testing.T) {
	c := NewClient()
	oo := OrdersOpts{}
	orders, _, err := c.Orders(oo)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("all", len(orders))
	for _, o := range orders {
		log.Println(o)
	}

}

func TestAddressImport(t *testing.T) {
	addr := "0x638ac149ea8ef9a1286c41b977017aa7359e6cfa"
	if A2T[addr] != "ALTS" {
		t.Fatal("wrong addr for ALTS")
	}
	if T2A["ZRX"] != "0xe41d2489571d322189246dafa5ebde1f4699f498" {
		t.Fatal("wrong ZRX addr")
	}
}

func TestOrderbook(t *testing.T) {
	c := NewClient()
	bt := "ZRX"
	qt := "WETH"
	lim := 10
	ob, _, err := c.Orderbook(OrderbookOpts{
		BaseTokenAddress:  T2A[bt],
		QuoteTokenAddress: T2A[qt],
	},
	)

	log.Printf("Book for %s/%s", bt, qt)
	log.Println("Base", bt, T2A[bt])
	log.Println("Quot", qt, T2A[qt])

	if err != nil {
		t.Fatal(err)
	}
	la := len(ob.Asks)
	lima := lim
	if la < lima {
		lima = la
	}
	log.Println(lima)
	for i := lima - 1; i >= 0; i = i - 1 {
		o := ob.Asks[i]
		bo, err := o.Process("Ask")
		if err != nil {
			t.Fatal(err)
		}
		log.Println(bo)
	}
	lb := len(ob.Bids)
	limb := lim
	if lb < limb {
		limb = lb
	}
	log.Println("Bids", lb)
	for _, o := range ob.Bids[0:limb] {
		bo, err := o.Process("Bid")
		if err != nil {
			t.Fatal(err)
		}
		log.Println(bo)
	}
}

func TestWSOrderbook(t *testing.T) {
	testPairEnvvar := "RRGO_TEST_PAIR"
	p := os.Getenv(testPairEnvvar)
	if len(p) == 0 {
		t.Fatalf("set %s", testPairEnvvar)
	}
	ts := strings.Split(p, "/")
	if len(ts) != 2 {
		t.Fatalf("%s must be in format <BaseToken>/<QuoteToken>", testPairEnvvar)
	}

	bt := ts[0]
	qt := ts[1]
	wso, err := NewWSOrderbook(T2A[bt], T2A[qt], snapshotLimit)
	if err != nil {
		t.Fatal(err)
	}
	wso.Run()
}
