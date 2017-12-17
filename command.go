package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type InboundMessage struct {
	To      []string    `json:"to"`
	From    string      `json:"from"`
	Type    string      `json:"type"`
	Command string      `json:"command"`
	Data    interface{} `json:"data"`
}

type OutboundMessage struct {
	From string      `json:"from"`
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func (c *Client) processCommand(message []byte, id string) error {
	in := &InboundMessage{}
	if err := json.Unmarshal(message, in); err != nil {
		return err
	}
	in.From = id
	if in.Command != "" {
		in.Command = strings.Title(strings.ToLower(in.Command))
		_, err := CallFuncByName(c, in.Command, in)
		return err
	} else {
		out := &OutboundMessage{
			From: in.From,
			Type: in.Type,
			Data: in.Data,
		}
		outMessage, err := json.Marshal(out)
		if err != nil {
			return err
		}
		for _, to := range in.To {
			for client := range c.hub.clients {
				if client.id == to {
					client.send <- outMessage
				}
			}
		}
		return nil
	}
}

func (c *Client) Broadcast(in *InboundMessage) {
	out := &OutboundMessage{
		From: in.From,
		Type: in.Type,
		Data: in.Data,
	}
	outMessage, err := json.Marshal(out)
	if err != nil {
		return
	}
	c.hub.broadcast <- outMessage
}

func CallFuncByName(myClass interface{}, funcName string, params ...interface{}) (out []reflect.Value, err error) {
	myClassValue := reflect.ValueOf(myClass)
	m := myClassValue.MethodByName(funcName)
	if !m.IsValid() {
		return make([]reflect.Value, 0), fmt.Errorf(`Method not found "%s"`, funcName)
	}
	in := make([]reflect.Value, len(params))
	for i, param := range params {
		in[i] = reflect.ValueOf(param)
	}
	out = m.Call(in)
	return
}
