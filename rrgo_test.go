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

	pr = PairsOpts{}
	pairs, _, err = c.Pairs(pr)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("all", len(pairs))
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
		t.Fatal("something wrong")
	}
}

func TestOrderbook(t *testing.T) {
	baseT := T2A["WETH"]
	quoteT := T2A["ZRX"]

	oo := OrderbookOpts{BaseTokenAddress: baseT, QuoteTokenAddress: quoteT}

	c := NewClient()
	ob, _, err := c.Orderbook(oo)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("Asks", len(ob.Asks))
	for _, o := range ob.Asks {
		log.Println(o)
	}
	log.Println("Bids", len(ob.Bids))
	for _, o := range ob.Bids {
		log.Println(o)
	}
}
