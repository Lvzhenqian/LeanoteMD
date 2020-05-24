package main


type Book struct {
	Id string`json:"id"`
	Title string`json:"title"`
	ParentId string`json:"pid"`
	Child []*Book`json:"child"`
}

func has(nodes []*Book, node *Book) (childs []*Book, ok bool) {
	for _, v:= range nodes{
		if v.ParentId == node.Id{
			childs = append(childs,v)
		}

	}

	if childs != nil {
		ok = true
	}
	return
}

func MakeTree(nodes []*Book, node *Book) {
	childs,_ := has(nodes,node)
	if childs != nil{
		node.Child = append(node.Child,childs[0:]...)
		for _,v := range childs{
			if _,ok := has(nodes,v);ok{
				MakeTree(nodes,v)
			}
		}
	}
}

