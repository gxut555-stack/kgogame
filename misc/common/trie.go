package common

import (
	"strings"
)

// 敏感词算法

// TrieNode 表示 Trie 树的节点
type TrieNode struct {
	children map[rune]*TrieNode
	isEnd    bool
}

// Trie 表示 Trie 树
type Trie struct {
	root *TrieNode
}

// NewTrie 创建一个新的 Trie 树
func NewTrie() *Trie {
	return &Trie{
		root: &TrieNode{
			children: make(map[rune]*TrieNode),
		},
	}
}

// Insert 将一个敏感词插入到 Trie 树中
func (t *Trie) Insert(word string) {
	current := t.root

	for _, ch := range word {
		node, ok := current.children[ch]
		if !ok {
			node = &TrieNode{
				children: make(map[rune]*TrieNode),
			}
			current.children[ch] = node
		}
		current = node
	}
	current.isEnd = true
}

// Contains 检查文本中是否包含敏感词
func (t *Trie) Contains(text string) bool {
	words := strings.Fields(text)
	for _, word := range words {
		if t.containsWord(word) {
			return true
		}
	}
	return false
}

// containsWord 检查单个词是否包含敏感词
func (t *Trie) containsWord(word string) bool {
	current := t.root

	for _, ch := range word {
		node, ok := current.children[ch]
		if !ok {
			return false
		}
		current = node
		if current.isEnd {
			return true
		}
	}
	return false
}
