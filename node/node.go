package node

type Node struct {
	ip string
	Cluster []Node
}

func New(ip string, cluster []Node) Node {
	return Node{
		ip:      ip,
		Cluster: cluster,
	}
}

func (n *Node) request(file string) {
	ready := make([]Node, 0)

	for _, node := range n.Cluster {
		if node.get(file){
			ready = append(ready, node)
		}
	}
}

// returns true if has the file
func (n *Node) get(file string) bool {

}

func (n *Node) discover() {
	for _, node := range n.Cluster {

	}
}
