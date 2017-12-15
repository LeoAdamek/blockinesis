package blockchain

import (
	"github.com/gorilla/websocket"
	"encoding/json"
)

// WebSocketURL is the URL of the websocket where we get our data
const WebSocketURL = "wss://ws.blockchain.info/inv"

type Client struct {
	ws *websocket.Conn
}

func New() (Client, error) {
	ws, _, err := websocket.DefaultDialer.Dial(WebSocketURL, nil)

	if err != nil {
		return Client{}, err
	}

	c := Client {
		ws: ws,
	}

	return c, err
}


func (c *Client) WatchTransactions(transactions chan Transaction, errors chan error) error {

	err := c.ws.WriteJSON(BasicRequest{Op: "unconfirmed_sub"})

	if err != nil {
		if errors != nil {
			errors <- err
		}
		return err
	}

	for {
		var data json.RawMessage

		resp := Response {
			Data: &data,
		}

		_, message, err := c.ws.ReadMessage()

		if err != nil && errors != nil {
				errors <- err
		}

		err = json.Unmarshal(message, &resp)

		if err != nil && errors != nil {
			errors <- err
		}

		if resp.Op == "utx" {
			var t Transaction

			err = json.Unmarshal(data, &t)

			if err != nil && errors != nil {
				errors <- err
			}

			if len(t.Inputs) > 0 {
				t.Value = t.Inputs[0].Out.Value
				t.ValueBTC = float64(t.Value) / 1e9
				
				t.From = t.Inputs[0].Out.Address
			}
			
			if len(t.Outputs) > 0 {
				t.To = t.Outputs[0].Address
			}
			
			
			transactions <- t
		}
	}
}