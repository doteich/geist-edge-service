package api

import (
	"context"
	"fmt"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/id"
	"github.com/gopcua/opcua/ua"
)

type BrowseInput struct {
	Body BrowseBody
}

type BrowseBody struct {
	UUID   string `json:"uuid" required:"true"`
	NodeID string `json:"nodeId" example:"ns=2;s=MyDevice" doc:"The NodeID to browse. Defaults to Objects folder if empty."`
}

type BrowseOutput struct {
	Body struct {
		Nodes []NodeInfo `json:"nodes"`
	}
}

type NodeInfo struct {
	NodeID      string `json:"nodeId"`
	BrowseName  string `json:"browseName"`
	DisplayName string `json:"displayName"`
	NodeClass   string `json:"nodeClass"`
	DataType    string `json:"dataType"`
}

func (a *AppState) Browse(ctx context.Context, input *BrowseInput) (*BrowseOutput, error) {
	mu.Lock()
	sess, ok := sessions[input.Body.UUID]
	if ok {
		sess.LastAccessed = time.Now()
	}
	mu.Unlock()

	if !ok {
		return nil, fmt.Errorf("session not found or expired")
	}

	nodeIDStr := input.Body.NodeID
	if nodeIDStr == "" {
		nodeIDStr = "i=85"
	}

	nodeId, err := ua.ParseNodeID(nodeIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid node id: %v", err)
	}

	node := sess.Client.Node(nodeId)

	var children []*opcua.Node

	results, err := node.ReferencedNodes(ctx, id.HasComponent, ua.BrowseDirectionForward, ua.NodeClassAll, true)
	if err != nil {
		return nil, fmt.Errorf("browse failed: %v", err)
	}

	children = append(children, results...)

	results, err = node.ReferencedNodes(ctx, id.Organizes, ua.BrowseDirectionForward, ua.NodeClassAll, true)
	if err != nil {
		return nil, fmt.Errorf("browse failed: %v", err)
	}

	children = append(children, results...)

	results, err = node.ReferencedNodes(ctx, id.HasProperty, ua.BrowseDirectionForward, ua.NodeClassAll, true)
	if err != nil {
		return nil, fmt.Errorf("browse failed: %v", err)
	}

	children = append(children, results...)

	resp := &BrowseOutput{}
	nodes := make([]NodeInfo, 0)

	for _, child := range children {

		attr, err := child.Attributes(ctx, ua.AttributeIDBrowseName, ua.AttributeIDDisplayName, ua.AttributeIDNodeClass, ua.AttributeIDDataType)
		if err != nil {
			continue
		}

		nc, err := child.NodeClass(ctx)

		if err != nil {
			continue
		}

		dt := "Object"

		if nc == ua.NodeClassVariable {
			dt = id.Name(attr[3].Value.NodeID().IntID())
		}

		nodes = append(nodes, NodeInfo{
			NodeID:      child.ID.String(),
			BrowseName:  attr[0].Value.String(),
			DisplayName: attr[1].Value.String(),
			NodeClass:   nc.String(),
			DataType:    dt,
		})
	}

	resp.Body.Nodes = nodes

	return resp, nil
}
