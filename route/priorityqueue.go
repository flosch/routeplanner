package main

import (
	"private/routenplaner/src/src/common"

	"code.google.com/p/go-priority-queue/prio"
)

type prioItem struct {
	priority int
	node     *common.Node
	prev     *prioItem
	hint     *HintMgr // Hinweise, wenn man von Prev auf Node dem Weg folgen möchte (z. B. "Könnte von Tor versperrt sein")
	index    int
}

type prioQueue struct {
	q prio.Queue
}

func (pi *prioItem) Less(y prio.Interface) bool {
	return pi.priority < y.(*prioItem).priority
}

func (pi *prioItem) Index(i int) {
	pi.index = i
}

func newPrioQueue() *prioQueue {
	return &prioQueue{}
}

func (pq *prioQueue) add(prev *prioItem, node *common.Node, priority int, hint *HintMgr) *prioItem {
	i := &prioItem{
		priority: priority,
		node:     node,
		prev:     prev,
		hint:     hint,
	}
	pq.q.Push(i)
	return i
}

func (pq *prioQueue) next() *prioItem {
	i := pq.q.Pop()
	return i.(*prioItem)
}

func (pq *prioQueue) Notifiy(i *prioItem) {
	pq.q.Fix(i.index)
}

func (pq *prioQueue) Len() int {
	return pq.q.Len()
}
