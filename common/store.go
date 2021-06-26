package common

//存储类型,用来表示文件存到哪里
type StoreType int

const (
	_ StoreType = iota
	//StoreLocal：节点本地
	StoreLocal
	//StoreCeph:Ceph集群
	StoreCeph
	//StoreOSS:阿里oss
	StoreOSS
	/*StoreMix:混合(ceph以及oss)*/
	StoreMix
	//StoreAll:所有类型的存储都存一份数据
	StoreAll
)
