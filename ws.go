package rrgo

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/buger/jsonparser"
	"github.com/gorilla/websocket"
)

var (
	websocketErrs = []int{
		websocket.CloseNormalClosure,
		websocket.CloseGoingAway,
		websocket.CloseProtocolError,
		websocket.CloseUnsupportedData,
		websocket.CloseNoStatusReceived,
		websocket.CloseAbnormalClosure,
		websocket.CloseInvalidFramePayloadData,
		websocket.ClosePolicyViolation,
		websocket.CloseMessageTooBig,
		websocket.CloseMandatoryExtension,
		websocket.CloseInternalServerErr,
		websocket.CloseServiceRestart,
		websocket.CloseTryAgainLater,
		websocket.CloseTLSHandshake,
	}
)

const (
	WSURL         = "wss://ws.radarrelay.com/0x/v0/ws"
	snapshotLimit = 20
)

type WSOrderbook struct {
	WS                 *websocket.Conn
	BaseTokenAddress   string
	QuoteTokenAddress  string
	Pair               string
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

func (wso *WSOrderbook) Subscribe(limit int) error {
	ws, err := openWebsocket()
	if err != nil {
		return err
	}
	rand.Seed(time.Now().UnixNano())
	rID := rand.Int() % 5096

	sm := SubscribeMessage{
		MessageFields: MessageFields{
			Type:      "subscribe",
			Channel:   "orderbook",
			RequestID: rID,
		},
		Payload: SubscribePayload{
			Snapshot:          true,
			Limit:             limit,
			BaseTokenAddress:  wso.BaseTokenAddress,
			QuoteTokenAddress: wso.QuoteTokenAddress,
		},
	}
	bsm, err := json.Marshal(sm)

	log.Println("Subscribing by", string(bsm))
	wso.WS = ws
	wso.SubscribeRequestID = rID
	err = wso.WS.WriteMessage(websocket.TextMessage, bsm)
	if err != nil {
		return err
	}
	return nil
}

func NewWSOrderbook(baseTA, quoteTA string, limit int) (*WSOrderbook, error) {
	pair := fmt.Sprintf("%s/%s", A2T[baseTA], A2T[quoteTA])
	log.Println("creating websocket for", pair)
	wso := &WSOrderbook{
		WS:                 nil,
		BaseTokenAddress:   baseTA,
		QuoteTokenAddress:  quoteTA,
		Pair:               pair,
		SubscribeRequestID: 0,
	}
	err := wso.Subscribe(limit)
	if err != nil {
		return nil, err
	}

	return wso, nil
}

type MessageFields struct {
	Type      string `json:"type"`
	Channel   string `json:"channel"`
	RequestID int    `json:"requestId"`
}

type SubscribePayload struct {
	BaseTokenAddress  string `json:"baseTokenAddress"`
	QuoteTokenAddress string `json:"quoteTokenAddress"`
	Snapshot          bool   `json:"snapshot"`
	Limit             int    `json:"limit"`
}

type SubscribeMessage struct {
	MessageFields
	Payload SubscribePayload `json:"payload"`
}

type SnapshotMessage struct {
	MessageFields
	Payload *Orderbook `json:"payload"`
}

type UpdateMessage struct {
	MessageFields
	Payload *APIOrder `json:"payload"`
}

type OfTheDayMessage struct {
	MOTD          string   `json:"motd"`
	Announcements []string `json:"announcements"`
}

func (wso *WSOrderbook) Run() {

	for {
		_, msg, err := wso.WS.ReadMessage()
		if err != nil {
			log.Println("ERROR:", err)
			if websocket.IsCloseError(err, websocketErrs...) {
				log.Println(wso.Pair, "re-opening Websocket")
				wso.WS.Close()
				err := wso.Subscribe(snapshotLimit)
				if err != nil {
					log.Fatal(err)
				}
				log.Println(wso.Pair, "Websocket re-opened")
			} else {
				log.Fatal(err)
			}
			continue
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
			snm.Payload.Reverse()
			log.Println(snm.Payload)
		case "update":
			um := UpdateMessage{}
			err := json.Unmarshal(msg, &um)
			if err != nil {
				log.Fatal()
			}
			bidAsk := "Bid"
			if um.Payload.MakerToken == wso.BaseTokenAddress {
				bidAsk = "Ask"
			}
			s, _ := um.Payload.Process(bidAsk)
			log.Printf("New %s: %s\n", bidAsk, s)
		default:
			motd := OfTheDayMessage{Announcements: []string{}}
			err := json.Unmarshal(msg, &motd)
			if err != nil {
				log.Println("ERROR WS receivedi garbage", string(msg))
			} else {
				log.Println("MOTD:", motd.MOTD)
				if len(motd.Announcements) > 0 {
					for _, a := range motd.Announcements {
						log.Println("*", a)
					}
				}
			}
		}

	}
	log.Println("Shouldnt be here")
}
