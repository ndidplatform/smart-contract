package db

import (
	"bytes"
	"fmt"
	"sync"
)

func init() {
	registerDBCreator(MemDBBackend, func(name string, dir string) (DB, error) {
		return NewMemDB(), nil
	}, false)
}

var _ DB = (*MemDB)(nil)

type MemDB struct {
	mtx sync.Mutex
	db  map[string][]byte
}

func NewMemDB() *MemDB {
	database := &MemDB{
		db: make(map[string][]byte),
	}
	return database
}

// Implements atomicSetDeleter.
func (db *MemDB) Mutex() *sync.Mutex {
	return &(db.mtx)
}

// Implements DB.
func (db *MemDB) Get(key []byte) []byte {
	db.mtx.Lock()
	defer db.mtx.Unlock()
	key = nonNilBytes(key)

	value := db.db[string(key)]
	return value
}

// Implements DB.
func (db *MemDB) Has(key []byte) bool {
	db.mtx.Lock()
	defer db.mtx.Unlock()
	key = nonNilBytes(key)

	_, ok := db.db[string(key)]
	return ok
}

// Implements DB.
func (db *MemDB) Set(key []byte, value []byte) {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	db.SetNoLock(key, value)
}

// Implements DB.
func (db *MemDB) SetSync(key []byte, value []byte) {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	db.SetNoLock(key, value)
}

// Implements atomicSetDeleter.
func (db *MemDB) SetNoLock(key []byte, value []byte) {
	db.SetNoLockSync(key, value)
}

// Implements atomicSetDeleter.
func (db *MemDB) SetNoLockSync(key []byte, value []byte) {
	key = nonNilBytes(key)
	value = nonNilBytes(value)

	db.db[string(key)] = value
	globalRoot.insert(key)
}

// Implements DB.
func (db *MemDB) Delete(key []byte) {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	db.DeleteNoLock(key)
}

// Implements DB.
func (db *MemDB) DeleteSync(key []byte) {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	db.DeleteNoLock(key)
}

// Implements atomicSetDeleter.
func (db *MemDB) DeleteNoLock(key []byte) {
	db.DeleteNoLockSync(key)
}

// Implements atomicSetDeleter.
func (db *MemDB) DeleteNoLockSync(key []byte) {
	key = nonNilBytes(key)

	delete(db.db, string(key))
	globalRoot.delete(key)
}

// Implements DB.
func (db *MemDB) Close() {
	// Close is a noop since for an in-memory
	// database, we don't have a destination
	// to flush contents to nor do we want
	// any data loss on invoking Close()
	// See the discussion in https://github.com/tendermint/tendermint/libs/pull/56
}

// Implements DB.
func (db *MemDB) Print() {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	for key, value := range db.db {
		fmt.Printf("[%X]:\t[%X]\n", []byte(key), value)
	}
}

// Implements DB.
func (db *MemDB) Stats() map[string]string {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	stats := make(map[string]string)
	stats["database.type"] = "memDB"
	stats["database.size"] = fmt.Sprintf("%d", len(db.db))
	return stats
}

// Implements DB.
func (db *MemDB) NewBatch() Batch {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	return &memBatch{db, nil}
}

//----------------------------------------
// Iterator

type avlTreeNode struct {
	parent     *avlTreeNode
	leftChild  *avlTreeNode
	rightChild *avlTreeNode
	height     int32
	key        []byte
}

func (node *avlTreeNode) getRoot() *avlTreeNode {
	for node.parent != nil {
		node = node.parent
	}
	return node
}

func displayTree(node *avlTreeNode) {
	fmt.Println("displayTree")
	/*if node == nil {
		return
	}
	fmt.Println("-", string(node.key))
	if node.leftChild != nil {
		fmt.Println("/", string(node.leftChild.key))
	}
	if node.rightChild != nil {
		fmt.Println("\\", string(node.rightChild.key))
	}
	displayTree(node.leftChild)
	displayTree(node.rightChild)
	if bytes.Equal(node.key, globalRoot.key) {
		fmt.Println("============================")
	}*/
}

//https://www.geeksforgeeks.org/avl-tree-set-1-insertion/

func leftRotate(x, y, z *avlTreeNode) {
	beyond := z.parent
	T3 := y.rightChild

	z.leftChild = T3
	if T3 != nil {
		T3.parent = z
	}

	y.rightChild = z
	z.parent = y

	y.parent = beyond
	if beyond != nil {
		if beyond.leftChild != nil {
			if bytes.Equal(beyond.leftChild.key, z.key) {
				beyond.leftChild = y
			} else {
				beyond.rightChild = y
			}
		} else {
			beyond.rightChild = y
		}
	}
}

func leftRightRotate(x, y, z *avlTreeNode) {
	beyond := z.parent
	T2 := x.leftChild
	T3 := x.rightChild

	//fmt.Println("LR", string(x.key), string(y.key), string(z.key), T2 == nil, T3 == nil)
	//displayTree(globalRoot)

	z.leftChild = T3
	if T3 != nil {
		T3.parent = z
	}

	y.rightChild = T2
	if T2 != nil {
		T2.parent = y
	}

	x.leftChild = y
	y.parent = x

	x.rightChild = z
	z.parent = x

	x.parent = beyond
	if beyond != nil {
		if beyond.leftChild != nil {
			if bytes.Equal(beyond.leftChild.key, z.key) {
				beyond.leftChild = x
			} else {
				beyond.rightChild = x
			}
		} else {
			beyond.rightChild = y
		}
	}
	//fmt.Println("After LR", string(x.key), string(y.key), string(z.key), T2 == nil, T3 == nil)
	//displayTree(globalRoot)
}

func rightRotate(x, y, z *avlTreeNode) {
	beyond := z.parent
	T2 := y.leftChild

	z.rightChild = T2
	if T2 != nil {
		T2.parent = z
	}

	y.leftChild = z
	z.parent = y

	y.parent = beyond
	if beyond != nil {
		if beyond.leftChild != nil {
			if bytes.Equal(beyond.leftChild.key, z.key) {
				beyond.leftChild = y
			} else {
				beyond.rightChild = y
			}
		} else {
			beyond.rightChild = y
		}
	}
}

func rightLeftRotate(x, y, z *avlTreeNode) {
	beyond := z.parent
	T2 := x.leftChild
	T3 := x.rightChild

	z.rightChild = T2
	if T2 != nil {
		T2.parent = z
	}

	y.leftChild = T3
	if T3 != nil {
		T3.parent = y
	}

	x.leftChild = z
	z.parent = x

	x.rightChild = y
	y.parent = x

	x.parent = beyond
	if beyond != nil {
		if beyond.leftChild != nil {
			if bytes.Equal(beyond.leftChild.key, z.key) {
				beyond.leftChild = x
			} else {
				beyond.rightChild = x
			}
		} else {
			beyond.rightChild = y
		}
	}
}

func reBalance(z, y, x *avlTreeNode) {
	fmt.Println("before reBalance")
	displayTree(globalRoot)
	code := 0
	if y.leftChild != nil {
		if bytes.Equal(y.leftChild.key, x.key) {
			code++
		}
	}
	if z.leftChild != nil {
		if bytes.Equal(z.leftChild.key, y.key) {
			code += 2
		}
	}
	fmt.Println("code", code, string(x.key), string(y.key), string(z.key))
	if code == 3 {
		leftRotate(x, y, z)
	} else if code == 2 {
		leftRightRotate(x, y, z)
	} else if code == 1 {
		rightLeftRotate(x, y, z)
	} else {
		rightRotate(x, y, z)
	}
	z.adjustHeight(true)
	if code == 3 || code == 0 {
		x.adjustHeight(true)
		y.adjustHeight(true)
	} else {
		y.adjustHeight(true)
		x.adjustHeight(true)
	}
	globalRoot = globalRoot.getRoot()
	//fmt.Println("After rebalance")
	fmt.Println("Global root after reBalance", string(globalRoot.key))
	displayTree(globalRoot)
}

func checkHeightDiff(path []*avlTreeNode) bool {
	fmt.Println("checkHeightDiff")
	node := path[len(path)-1]
	leftHeight := int32(-1)
	rightHeight := int32(-1)
	needRebalance := false
	if node.leftChild != nil {
		leftHeight = node.leftChild.height
	}
	if node.rightChild != nil {
		rightHeight = node.rightChild.height
	}
	if leftHeight > rightHeight+1 || rightHeight > leftHeight+1 {
		fmt.Println(leftHeight, rightHeight)
		reBalance(node, path[len(path)-2], path[len(path)-3])
		needRebalance = true
	}
	fmt.Println("Done checkHeightDiff")
	return needRebalance
}

func (node *avlTreeNode) adjustHeight(oneTine bool) {
	fmt.Println("adjustHeight")
	adjustingNode := node
	path := [](*avlTreeNode){}
	for adjustingNode != nil {
		path = append(path, adjustingNode)
		leftHeight := int32(-1)
		rightHeight := int32(-1)
		if adjustingNode.leftChild != nil {
			leftHeight = adjustingNode.leftChild.height
		}
		if adjustingNode.rightChild != nil {
			rightHeight = adjustingNode.rightChild.height
		}
		if leftHeight > rightHeight {
			adjustingNode.height = leftHeight + 1
		} else {
			adjustingNode.height = rightHeight + 1
		}
		if oneTine {
			return
		}
		adjustingNode = adjustingNode.parent
		if checkHeightDiff(path) {
			return
		}
	}
}

func (node *avlTreeNode) findExactOrClosetLeafNode(key []byte) (bool, *avlTreeNode) {
	if bytes.Equal(key, node.key) {
		return true, node
	} else if bytes.Compare(key, node.key) == -1 {
		if node.leftChild == nil {
			return false, node
		}
		return node.leftChild.findExactOrClosetLeafNode(key)
	} else {
		if node.rightChild == nil {
			return false, node
		}
		return node.rightChild.findExactOrClosetLeafNode(key)
	}
}

func (node *avlTreeNode) getNext(cur []byte, reverse bool) []byte {
	root := node.getRoot()
	exact, targetNode := root.findExactOrClosetLeafNode(cur)
	if !exact {
		panic("This should not happen, invalid iterator")
	}
	if reverse {
		next := targetNode.leftChild
		if next == nil {
			parent := targetNode.parent
			next = targetNode
			for bytes.Equal(parent.leftChild.key, next.key) {
				next = parent
				parent = parent.parent
			}
			return parent.key
		}
		for next.rightChild != nil {
			next = next.rightChild
		}
		return next.key
	}
	next := targetNode.rightChild
	if next == nil {
		parent := targetNode.parent
		next = targetNode
		for bytes.Equal(parent.rightChild.key, next.key) {
			next = parent
			parent = parent.parent
		}
		return parent.key
	}
	for next.leftChild != nil {
		next = next.leftChild
	}
	return next.key
}

func (node *avlTreeNode) insert(key []byte) {
	fmt.Println("insert", string(key))
	root := node.getRoot()
	if root.key == nil {
		root.key = key
	} else {
		exact, closetNode := root.findExactOrClosetLeafNode(key)
		if !exact {
			fmt.Println("Closet Node", string(closetNode.key))
			newNode := &avlTreeNode{
				parent:     nil,
				leftChild:  nil,
				rightChild: nil,
				height:     0,
				key:        key,
			}
			if bytes.Compare(key, closetNode.key) == -1 {
				closetNode.leftChild = newNode
				newNode.parent = closetNode
			} else {
				closetNode.rightChild = newNode
				newNode.parent = closetNode
			}
			newNode.adjustHeight(false)
		} else {
			closetNode.key = key
			fmt.Println("Update avl unchanged")
		}
	}
	displayTree(globalRoot)
}

func (node *avlTreeNode) delete(key []byte) {
	fmt.Println("delete", string(key))
	root := node.getRoot()
	if root.key != nil {
		exact, targetNode := root.findExactOrClosetLeafNode(key)
		if exact {
			if targetNode.leftChild == nil || targetNode.rightChild == nil {
				parent := targetNode.parent
				var newChild *avlTreeNode
				if targetNode.leftChild == nil {
					newChild = targetNode.rightChild
				} else if targetNode.rightChild == nil {
					newChild = targetNode.leftChild
				}
				if parent != nil {
					if bytes.Equal(parent.leftChild.key, key) {
						parent.leftChild = newChild
					} else {
						parent.rightChild = newChild
					}
					if newChild != nil {
						newChild.parent = parent
					}
					parent.adjustHeight(false)
				} else {
					globalRoot = newChild
					newChild.parent = nil
				}
			} else {
				scapeGoat := targetNode.leftChild
				for scapeGoat.rightChild != nil {
					scapeGoat = scapeGoat.rightChild
				}
				newKey := scapeGoat.key
				root.delete(newKey)
				targetNode.key = newKey
			}
			displayTree(globalRoot)
		} else {
			fmt.Println("Unfound key")
		}
	}
}

var globalRoot = &avlTreeNode{
	parent:     nil,
	leftChild:  nil,
	rightChild: nil,
	height:     0,
	key:        nil,
}

// We need a copy of all of the keys.
// Not the best, but probably not a bottleneck depending.
type memDBIterator struct {
	db      DB
	cur     []byte
	root    *avlTreeNode
	start   []byte
	end     []byte
	reverse bool
}

var _ Iterator = (*memDBIterator)(nil)

// Implements DB.
func (db *MemDB) Iterator(start, end []byte) Iterator {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	return newMemDBIterator(db, start, end, false)
}

// Implements DB.
func (db *MemDB) ReverseIterator(start, end []byte) Iterator {
	db.mtx.Lock()
	defer db.mtx.Unlock()

	return newMemDBIterator(db, start, end, true)
}

// Keys is expected to be in reverse order for reverse iterators.
func newMemDBIterator(db DB, start, end []byte, reverse bool) *memDBIterator {
	return &memDBIterator{
		db:      db,
		cur:     start,
		root:    globalRoot,
		start:   start,
		end:     end,
		reverse: reverse,
	}
}

// Implements Iterator.
func (itr *memDBIterator) Domain() ([]byte, []byte) {
	return itr.start, itr.end
}

// Implements Iterator.
func (itr *memDBIterator) Valid() bool {
	exact, _ := itr.root.findExactOrClosetLeafNode(itr.cur)
	return exact && IsKeyInDomain([]byte(itr.cur), itr.start, itr.end, itr.reverse)
}

// Implements Iterator.
func (itr *memDBIterator) Next() {
	itr.assertIsValid()
	itr.cur = itr.root.getNext(itr.cur, itr.reverse)
}

// Implements Iterator.
func (itr *memDBIterator) Key() []byte {
	itr.assertIsValid()
	return itr.cur
}

// Implements Iterator.
func (itr *memDBIterator) Value() []byte {
	itr.assertIsValid()
	key := itr.Key()
	return itr.db.Get(key)
}

// Implements Iterator.
func (itr *memDBIterator) Close() {
	itr.root = nil
	itr.db = nil
}

func (itr *memDBIterator) assertIsValid() {
	if !itr.Valid() {
		panic("memDBIterator is invalid")
	}
}

//----------------------------------------
// Misc.

/*func (db *MemDB) getSortedKeys(start, end []byte, reverse bool) []string {
	keys := []string{}
	for key := range db.db {
		inDomain := IsKeyInDomain([]byte(key), start, end, reverse)
		if inDomain {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	if reverse {
		nkeys := len(keys)
		for i := 0; i < nkeys/2; i++ {
			temp := keys[i]
			keys[i] = keys[nkeys-i-1]
			keys[nkeys-i-1] = temp
		}
	}
	return keys
}*/
