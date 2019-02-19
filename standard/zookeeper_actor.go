package standard

import (
	"errors"
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	. "github.com/xxpxxxxp/goactor/system"
	"path"
	"reflect"
)

type CreateNodeRequest struct {
	Path      string
	Recursive bool
}

type CreateNodeResponse struct {
	Path  string
	Error error
}

type NodeOperation int

const (
	NodeCreate NodeOperation = iota
	NodeDelete
)

type BatchNodesOperationRequest struct {
	Paths     []string
	Operation NodeOperation
}

type WatchPathRequest struct {
	Caller string
	Path   string
}

type RemoveNodeRequest string
type RmrRequest string
type GetNodeDataRequest string

type GetNodeDataResponse struct {
	Data  string
	Error error
}

type SetNodeDataRequest struct {
	Path string
	Data string
}

type GetSubNodesRequest string

type GetSubNodesResponse struct {
	SubNodes []string
	Error    error
}

type ZookeeperActor struct {
	Conn     *zk.Conn
	BasePath string
}

func (zoo *ZookeeperActor) Receive(system *ActorSystem, eventType EventType, event interface{}) interface{} {
	switch request := event.(type) {
	case CreateNodeRequest:
		var p string
		var err error
		if request.Recursive {
			p, err = createZNodeRecursive(zoo.Conn, request.Path)
			return &CreateNodeResponse{p, err}
		} else {
			p, err = zoo.Conn.Create(request.Path, []byte(nil), int32(0), zk.WorldACL(zk.PermAll))
		}
		return &CreateNodeResponse{p, err}
	case BatchNodesOperationRequest:
		switch request.Operation {
		case NodeCreate:
			return bulkCreateZNodes(zoo.Conn, request.Paths)
		case NodeDelete:
			return bulkDeleteZNodes(zoo.Conn, request.Paths)
		}
	case RemoveNodeRequest:
		return removeZNode(zoo.Conn, string(request))
	case RmrRequest:
		return rmrZNode(zoo.Conn, string(request))
	case GetNodeDataRequest:
		data, _, err := zoo.Conn.Get(string(request))
		result := ""
		if err == nil {
			result = string(data)
		}
		return &GetNodeDataResponse{result, err}
	case SetNodeDataRequest:
		_, err := zoo.Conn.Set(request.Path, []byte(request.Data), -1)
		return err
	case GetSubNodesRequest:
		subNodes, _, err := zoo.Conn.Children(string(request))
		return &GetSubNodesResponse{subNodes, err}
	default:
		return errors.New(fmt.Sprintf("unsupported event type \"%s\" for ZookeeperActor", reflect.TypeOf(request).Name()))
	}

	return nil
}

func createZNodeRecursive(conn *zk.Conn, node string) (string, error) {
	if exist, _, err := conn.Exists(node); err != nil {
		return node, err
	} else if exist {
		return node, nil
	}

	if p, err := createZNodeRecursive(conn, path.Dir(node)); err != nil && err != zk.ErrNodeExists {
		return p, err
	}

	return conn.Create(node, []byte(nil), int32(0), zk.WorldACL(zk.PermAll))
}

func bulkCreateZNodes(conn *zk.Conn, nodes []string) error {
	createRequests := make([]interface{}, len(nodes))
	for i, node := range nodes {
		createRequests[i] = &zk.CreateRequest{Path: node, Data: nil, Acl: zk.WorldACL(zk.PermAll), Flags: int32(0)}
	}
	if _, err := conn.Multi(createRequests...); err != nil {
		return err
	}
	return nil
}

func bulkDeleteZNodes(conn *zk.Conn, nodes []string) error {
	deleteRequests := make([]interface{}, len(nodes))
	for i, node := range nodes {
		deleteRequests[i] = &zk.DeleteRequest{Path: node, Version: -1}
	}
	if _, err := conn.Multi(deleteRequests...); err != nil {
		return err
	}
	return nil
}

func removeZNode(conn *zk.Conn, node string) error {
	if exist, _, err := conn.Exists(node); err != nil {
		return err
	} else if !exist {
		return errors.New("node not exist") // path not exist
	}

	return conn.Delete(node, int32(-1))
}

func rmrZNode(conn *zk.Conn, root string) error {
	if exist, _, err := conn.Exists(root); err != nil {
		return err
	} else if exist {
		// rmr sub tree DFS
		if children, _, err := conn.Children(root); err == nil && len(children) > 0 {
			childrenFullPath := make([]string, len(children))
			for i, child := range children {
				childrenFullPath[i] = fmt.Sprintf("%s/%s", root, child)
				if err := rmrZNode(conn, childrenFullPath[i]); err != nil {
					return err
				}
			}
			if err := bulkDeleteZNodes(conn, childrenFullPath); err != nil {
				return err
			}
		}
		if err := conn.Delete(root, -1); err != nil && err != zk.ErrNoNode {
			return err
		}
	}
	return nil
}
