package main

import "fmt"

// copyOnWriteContext 模拟 google/btree 中的 ownership 上下文。
// 只有 node.cow == tree.cow 时，节点才能被当前树原地修改。
type copyOnWriteContext struct {
	tag string
}

type node struct {
	key   int
	value int
	left  *node
	right *node
	cow   *copyOnWriteContext
}

// mutableFor 与 google/btree 的思路一致：
// 若节点不归当前 tree.cow 所有，则复制一个可写副本（children 指针先共享）。
func (n *node) mutableFor(cow *copyOnWriteContext) *node {
	if n == nil {
		return nil
	}
	if n.cow == cow {
		return n
	}
	out := *n
	out.cow = cow
	return &out
}

type Tree struct {
	root *node
	cow  *copyOnWriteContext
}

func NewTree(tag string) *Tree {
	return &Tree{cow: &copyOnWriteContext{tag: tag}}
}

// Clone 模仿 google/btree：
// 1. 保留旧节点（旧 cow）作为共享只读结构
// 2. 给两棵树各分配一个新的 cow 上下文
// 3. 后续任意一方写入时触发路径复制（COW）
func (t *Tree) Clone() *Tree {
	cow1, cow2 := *t.cow, *t.cow
	out := *t
	t.cow = &cow1
	out.cow = &cow2
	return &out
}

func (t *Tree) Upsert(key, value int) {
	if t.root == nil {
		t.root = &node{key: key, value: value, cow: t.cow}
		return
	}
	t.root = t.root.mutableFor(t.cow)
	upsertNode(t.root, key, value, t.cow)
}

func upsertNode(n *node, key, value int, cow *copyOnWriteContext) {
	switch {
	case key < n.key:
		if n.left == nil {
			n.left = &node{key: key, value: value, cow: cow}
			return
		}
		n.left = n.left.mutableFor(cow)
		upsertNode(n.left, key, value, cow)
	case key > n.key:
		if n.right == nil {
			n.right = &node{key: key, value: value, cow: cow}
			return
		}
		n.right = n.right.mutableFor(cow)
		upsertNode(n.right, key, value, cow)
	default:
		n.value = value
	}
}

func (t *Tree) Get(key int) (int, bool) {
	n := t.root
	for n != nil {
		switch {
		case key < n.key:
			n = n.left
		case key > n.key:
			n = n.right
		default:
			return n.value, true
		}
	}
	return 0, false
}

func ptr(n *node) string {
	if n == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%p", n)
}

func printPointers(name string, t *Tree) {
	left := (*node)(nil)
	right := (*node)(nil)
	if t.root != nil {
		left = t.root.left
		right = t.root.right
	}
	fmt.Printf("%s(cowTag=%s cowPtr=%p): root=%s rootCow=%p left=%s right=%s\n",
		name, t.cow.tag, t.cow, ptr(t.root), t.root.cow, ptr(left), ptr(right))
}

func main() {
	origin := NewTree("origin-cow")
	origin.Upsert(10, 100)
	origin.Upsert(5, 50)
	origin.Upsert(15, 150)

	clone := origin.Clone()
	fmt.Println("== 1) Clone 后：两棵树共享旧节点（只读共享）==")
	printPointers("origin", origin)
	printPointers("clone ", clone)
	fmt.Printf("共享 root? %v\n\n", origin.root == clone.root)

	fmt.Println("== 2) 写 clone：只复制写路径（root + left），right 继续共享 ==")
	clone.Upsert(5, 500)
	printPointers("origin", origin)
	printPointers("clone ", clone)
	fmt.Printf("root 仍共享? %v\n", origin.root == clone.root)
	fmt.Printf("left 仍共享? %v\n", origin.root.left == clone.root.left)
	fmt.Printf("right 仍共享? %v\n", origin.root.right == clone.root.right)
	v1, _ := origin.Get(5)
	v2, _ := clone.Get(5)
	fmt.Printf("value(5): origin=%d, clone=%d\n\n", v1, v2)

	fmt.Println("== 3) 再写 origin 右子树：继续各自复制自己的写路径 ==")
	origin.Upsert(15, 1500)
	printPointers("origin", origin)
	printPointers("clone ", clone)
	v3, _ := origin.Get(15)
	v4, _ := clone.Get(15)
	fmt.Printf("value(15): origin=%d, clone=%d\n", v3, v4)

	type pair struct {
		key   int
		value int
	}

	var p = new(pair)
	p.key = 1
	p.value = 100

	p1, p2 := *p, *p
	if p1 == p2 {
		fmt.Println("cp1 == cp2")
	} else {
		fmt.Println("cp1 != cp2")
	}

	pp1, pp2 := p, p
	if pp1 == pp2 {
		fmt.Println("pp1 == pp2")
	} else {
		fmt.Println("pp1 != pp2")
	}

	cp1, cp2 := &p1, &p2
	if cp1 == cp2 {
		fmt.Println("cp1 == cp2")
	} else {
		fmt.Println("cp1 != cp2")
	}

}
