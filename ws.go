package rrgo

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/buger/jsonparser"
	"github.com/gorilla/websocket"
)

const WSURL = "wss://ws.radarrelay.com/0x/v0/ws"

type WSOrderbook struct {
	WS                 *websocket.Conn
	BaseTokenAddress   string
	QuoteTokenAddress  string
	SubscribeRequestID int
}

func openWebsocket() (*websocket.Conn, error) {
	dialer := websocket.Dialer{
		HandshakeTimeout: time.Second * 5,
	}
	c, _, err := dialer.Dial(WSURL, nil)
	if err != nil {
		return nil, err
	}
	return c, nil

}

func NewWSOrderbook(baseTA, quoteTA string, limit int) (*WSOrderbook, error) {
	log.Printf("creating websocket for %s/%s\n", A2T[baseTA], A2T[quoteTA])

	ws, err := openWebsocket()
	if err != nil {
		return nil, err
	}
	rand.Seed(time.Now().UnixNano())
	wso := WSOrderbook{
		WS:                 ws,
		BaseTokenAddress:   baseTA,
		QuoteTokenAddress:  quoteTA,
		SubscribeRequestID: rand.Int(),
	}

	sm := SubscribeMessage{
		Type:      "subscribe",
		Channel:   "orderbook",
		RequestID: rand.Int(),
		Payload: SubscribePayload{
			Snapshot:          true,
			Limit:             limit,
			BaseTokenAddress:  baseTA,
			QuoteTokenAddress: quoteTA,
		},
	}
	bsm, err := json.Marshal(sm)

	log.Println("subsMsg", string(bsm))
	err = wso.WS.WriteMessage(websocket.TextMessage, bsm)
	if err != nil {
		return nil, err
	}
	return &wso, nil
}

type SubscribePayload struct {
	BaseTokenAddress  string `json:"baseTokenAddress"`
	QuoteTokenAddress string `json:"quoteTokenAddress"`
	Snapshot          bool   `json:"snapshot"`
	Limit             int    `json:"limit"`
}

type SubscribeMessage struct {
	Type      string           `json:"type"`
	Channel   string           `json:"channel"`
	RequestID int              `json:"requestId"`
	Payload   SubscribePayload `json:"payload"`
}

type SnapshotMessage struct {
	Type      string     `json:"type"`
	Channel   string     `json:"channel"`
	RequestID int        `json:"requestId"`
	Payload   *Orderbook `json:"payload"`
}

func (wso *WSOrderbook) Run() {

	for {
		_, msg, err := wso.WS.ReadMessage()
		if err != nil {
			log.Fatal(err)
		}
		mtype, _ := jsonparser.GetUnsafeString(msg, "type")
		switch mtype {
		case "subscribe":
			sm := SubscribeMessage{}
			err := json.Unmarshal(msg, &sm)
			if err != nil {
				log.Fatal(err)
			}
			log.Println(sm)
		case "snapshot":
			snm := SnapshotMessage{}
			err := json.Unmarshal(msg, &snm)
			if err != nil {
				log.Fatal()
			}
			log.Println(snm.Payload)
		default:
			log.Println("RECEIVED", string(msg))
		}

	}
	log.Println("Shouldnt be here")
}
