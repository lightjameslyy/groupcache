/*
Copyright 2013 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package lru implements an LRU cache.
package lru

import (
	"container/list"
	"fmt"
)

// Cache is an LRU cache. It is not safe for concurrent access.
type Cache struct {
	// MaxEntries is the maximum number of cache entries before
	// an item is evicted. Zero means no limit.
	// lru容量限制，0表示无限制
	MaxEntries int

	// OnEvicted optionally specificies a callback function to be
	// executed when an entry is purged from the cache.
	// entry从cache中移出时的回调函数
	OnEvicted func(key Key, value interface{})

	// 辅助链表
	ll    *list.List
	// 存储cache数据，这里list.Element.Value的类型是*entry
	cache map[interface{}]*list.Element
}

// A Key may be any value that is comparable. See http://golang.org/ref/spec#Comparison_operators
type Key interface{}

type entry struct {
	key   Key
	value interface{}
}

// New creates a new Cache.
// If maxEntries is zero, the cache has no limit and it's assumed
// that eviction is done by the caller.
func New(maxEntries int) *Cache {
	return &Cache{
		MaxEntries: maxEntries,
		ll:         list.New(),
		cache:      make(map[interface{}]*list.Element),
	}
}

// Add adds a value to the cache.
// 向cache中添加entry
func (c *Cache) Add(key Key, value interface{}) {
	// 如果cache为空，先new出来
	if c.cache == nil {
		c.cache = make(map[interface{}]*list.Element)
		c.ll = list.New()
	}

	// 如果entry已存在，移到ll的最前面，更新value
	if ee, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ee)
		ee.Value.(*entry).value = value
		return
	}
	// 如果是新的entry，插入最前面
	ele := c.ll.PushFront(&entry{key, value})
	c.cache[key] = ele
	// 如果ll长度超过最大限制，删除最旧的entry
	if c.MaxEntries != 0 && c.ll.Len() > c.MaxEntries {
		c.RemoveOldest()
	}
}

// Get looks up a key's value from the cache.
// 查询key对应的entry的value
func (c *Cache) Get(key Key) (value interface{}, ok bool) {
	// 如果cache为空，返回默认值
	if c.cache == nil {
		return
	}

	// 如果命中，将entry放到最前面，返回entry的value
	if ele, hit := c.cache[key]; hit {
		c.ll.MoveToFront(ele)
		return ele.Value.(*entry).value, true
	}
	// 未命中，返回默认值
	return
}

// Remove removes the provided key from the cache.
func (c *Cache) Remove(key Key) {
	// 如果cache为空，返回
	if c.cache == nil {
		return
	}
	// 如果有对应的entry，将它删除
	if ele, hit := c.cache[key]; hit {
		c.removeElement(ele)
	}
}

// RemoveOldest removes the oldest item from the cache.
func (c *Cache) RemoveOldest() {
	// 如果cache为空，返回
	if c.cache == nil {
		return
	}
	// 从尾部删除
	ele := c.ll.Back()
	if ele != nil {
		c.removeElement(ele)
	}
}

func (c *Cache) removeElement(e *list.Element) {
	// 删除ll中的element
	c.ll.Remove(e)
	kv := e.Value.(*entry)
	// 删除map中对应的键值对
	delete(c.cache, kv.key)
	if c.OnEvicted != nil {
		// 调用回调函数
		c.OnEvicted(kv.key, kv.value)
	}
}

// Len returns the number of items in the cache.
func (c *Cache) Len() int {
	if c.cache == nil {
		return 0
	}
	return c.ll.Len()
}

// Clear purges all stored items from the cache.
func (c *Cache) Clear() {
	if c.OnEvicted != nil {
		for _, e := range c.cache {
			kv := e.Value.(*entry)
			c.OnEvicted(kv.key, kv.value)
		}
	}
	c.ll = nil
	c.cache = nil
}

func (c *Cache) String() (s string) {
	for k, v := range c.cache {
		s += fmt.Sprintf("key: %v, element: {key: %v, value: %v}\n", k,
			v.Value.(*entry).key, v.Value.(*entry).value)
	}
	return
}