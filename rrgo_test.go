package rrgo

import (
	"log"
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
	bt := "WETH"
	tt := "ZRX"
	ob, _, err := c.Orderbook(OrderbookOpts{
		BaseTokenAddress:  T2A[bt],
		QuoteTokenAddress: T2A[tt],
	},
	)

	log.Printf("Book for %s/%s, i.e. Base: %s, Quote: %s\n",
		tt, bt, bt, tt)

	if err != nil {
		t.Fatal(err)
	}
	la := len(ob.Asks)
	log.Println("Asks", la)
	for _, o := range ob.Asks[0:10] {
		log.Println(&o)
	}
	lb := len(ob.Bids)
	log.Println("Bids", lb)
	for _, o := range ob.Bids[0:10] {
		log.Println(&o)
	}
}
