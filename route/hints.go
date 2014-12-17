package main

import (
	"private/routenplaner/src/src/common"
)

type Hint struct {
	Node *common.Node
	Text string
}

func newHint(node *common.Node, text string) *Hint {
	return &Hint{
		Node: node,
		Text: text,
	}
}

type HintMgr struct {
	way        *common.Way
	node_hints map[*common.Node]string
	way_hints  map[string]*Hint
}

func newHintMgr(way *common.Way) *HintMgr {
	return &HintMgr{
		way:        way,
		node_hints: make(map[*common.Node]string),
		way_hints:  make(map[string]*Hint),
	}
}

func (hm *HintMgr) AddNodeHint(node *common.Node, text string) {
	hm.node_hints[node] = text
}

func (hm *HintMgr) AddWayHint(node *common.Node, text string) {
	hm.way_hints[text] = &Hint{
		Node: node,
		Text: text,
	}
}

func (hm *HintMgr) GetHints() []*Hint {
	result := make([]*Hint, 0, len(hm.node_hints)+len(hm.way_hints))

	for node, text := range hm.node_hints {
		result = append(result, &Hint{
			Node: node,
			Text: text,
		})
	}

	for _, hint := range hm.way_hints {
		result = append(result, hint)
	}

	return result
}

func (hm *HintMgr) Merge(hm2 *HintMgr) {
	for node, text := range hm2.node_hints {
		hm.node_hints[node] = text
	}

	for text, hint := range hm2.way_hints {
		if _, has_same_hint_for_way := hm.way_hints[text]; !has_same_hint_for_way {
			hm.way_hints[text] = hint
		}
	}

}
