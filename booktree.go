package main

type Book struct {
	Id string`json:"id"`
	Title string`json:"title"`
	ParentId string`json:"pid"`
	Child []*Book`json:"child"`
}

func hasChild(SubNodeList []*Book, node *Book) (childs []*Book, ok bool) {
	for _, v:= range SubNodeList{
		if v.ParentId == node.Id{
			childs = append(childs,v)
		}
	}
	if childs != nil {
		ok = true
	}
	return
}

func MakeTree(nodes []*Book, ParentNode *Book) {
	if childs,ok  := hasChild(nodes,ParentNode); ok {
		ParentNode.Child = append(ParentNode.Child,childs...) // 把存在的child列表关联到parent里
		// 递归子节点列表，查看是否还有下级子节点
		for _,v := range childs{
			if _,ok := hasChild(nodes,v);ok{
				MakeTree(nodes,v)
			}
		}
	}
}