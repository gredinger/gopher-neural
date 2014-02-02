package engine

import (
	"github.com/NOX73/go-neural"
	"github.com/NOX73/go-neural/lern"
	"github.com/NOX73/go-neural/persist"
)

const (
	lernChannelCapacity = 5
	calcChannelCapacity = 5
	dumpChannelCapacity = 5
)

type Engine interface {
	Start()
	Lern(in, ideal []float64, speed float64)
	Calculate([]float64) []float64
	Dump() *persist.NetworkDump
}

type engine struct {
	Network          *neural.Network
	LernChannel      chan *request
	CalculateChannel chan *request
	DumpChannel      chan *request
}

type request []interface{}

func New(n *neural.Network) Engine {
	e := &engine{
		Network:          n,
		LernChannel:      make(chan *request, lernChannelCapacity),
		CalculateChannel: make(chan *request, calcChannelCapacity),
		DumpChannel:      make(chan *request, dumpChannelCapacity),
	}

	return e
}

func (e *engine) Start() {
	go e.loop()
}

func (e *engine) Lern(in, ideal []float64, speed float64) {
	e.LernChannel <- &request{&in, &ideal, speed}
}

func (e *engine) Calculate(in []float64) []float64 {
	resp := make(chan *[]float64, 1)
	e.CalculateChannel <- &request{&in, resp}
	return *(<-resp)
}

func (e *engine) Dump() *persist.NetworkDump {
	resp := make(chan *persist.NetworkDump, 1)
	e.DumpChannel <- &request{resp}
	return <-resp
}

func (e *engine) loop() {
	for {

		select {
		case r := <-e.CalculateChannel:
			e.calculate(r)
		case r := <-e.DumpChannel:
			e.dump(r)
		default:
		}

		select {
		case r := <-e.DumpChannel:
			e.dump(r)
		case r := <-e.CalculateChannel:
			e.calculate(r)
		case r := <-e.LernChannel:
			e.lern(r)
		}

	}
}

func (e *engine) lern(req *request) {
	r := *req

	in := r[0].(*[]float64)
	ideal := r[1].(*[]float64)
	speed := r[2].(float64)
	lern.Lern(e.Network, *in, *ideal, speed)
}

func (e *engine) calculate(req *request) {
	r := *req

	in := r[0].(*[]float64)
	resp := r[1].(chan *[]float64)

	res := e.Network.Calculate(*in)
	resp <- &(res)
}

func (e *engine) dump(req *request) {
	r := *req
	resp := r[0].(chan *persist.NetworkDump)

	resp <- persist.ToDump(e.Network)
}
