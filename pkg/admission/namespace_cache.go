/*
 Licensed to the Apache Software Foundation (ASF) under one
 or more contributor license agreements.  See the NOTICE file
 distributed with this work for additional information
 regarding copyright ownership.  The ASF licenses this file
 to you under the Apache License, Version 2.0 (the
 "License"); you may not use this file except in compliance
 with the License.  You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package admission

import (
	"sync"

	v1 "k8s.io/api/core/v1"
	informersv1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/apache/yunikorn-k8shim/pkg/common/constants"
	"github.com/apache/yunikorn-k8shim/pkg/log"
)

type NamespaceCache struct {
	nameSpaces map[string]nsFlags

	sync.RWMutex
}

// nsFlags defines the two flags that can be set on the namespace.
// It needs to support a tri-state value showing presence besides true/false.
// -1: not present
// 0: false
// 1: true
type nsFlags struct {
	enableYuniKorn int
	generateAppID  int
}

// NewNamespaceCache creates a new cache and registers the handler for the cache with the Informer.
func NewNamespaceCache(namespaces informersv1.NamespaceInformer) *NamespaceCache {
	nsc := &NamespaceCache{
		nameSpaces: make(map[string]nsFlags),
	}
	if namespaces != nil {
		namespaces.Informer().AddEventHandler(&namespaceUpdateHandler{cache: nsc})
	}
	return nsc
}

// enableYuniKorn returns the value for the enableYuniKorn flag (tri-state -1, 0 or 1) for the namespace.
func (nsc *NamespaceCache) enableYuniKorn(name string) int {
	nsc.RLock()
	defer nsc.RUnlock()

	flag, ok := nsc.nameSpaces[name]
	if !ok {
		return -1
	}
	return flag.enableYuniKorn
}

// generateAppID returns the value for the generateAppID flag (tri-state -1, 0 or 1) for the namespace.
func (nsc *NamespaceCache) generateAppID(name string) int {
	nsc.RLock()
	defer nsc.RUnlock()

	flag, ok := nsc.nameSpaces[name]
	if !ok {
		return -1
	}
	return flag.generateAppID
}

// namespaceExists for test only to see if the namespace has been added to the cache or not.
func (nsc *NamespaceCache) namespaceExists(name string) bool {
	nsc.RLock()
	defer nsc.RUnlock()

	_, ok := nsc.nameSpaces[name]
	return ok
}

// namespaceUpdateHandler implements the K8s ResourceEventHandler interface for namespaces.
type namespaceUpdateHandler struct {
	cache *NamespaceCache
}

// OnAdd adds or replaces the namespace entry in the cache.
// The cached value is only the resulting value of the annotation, not the whole namespace object.
// An empty string for the Name is technically possible but should not occur.
func (h *namespaceUpdateHandler) OnAdd(obj interface{}) {
	ns := convert2Namespace(obj)
	if ns == nil {
		return
	}

	newFlags := getAnnotationValues(ns)
	h.cache.Lock()
	defer h.cache.Unlock()
	h.cache.nameSpaces[ns.Name] = newFlags
}

// OnUpdate calls OnAdd for processing the namespace cache update.
func (h *namespaceUpdateHandler) OnUpdate(_, newObj interface{}) {
	h.OnAdd(newObj)
}

// OnDelete removes the namespace from the cache.
func (h *namespaceUpdateHandler) OnDelete(obj interface{}) {
	var ns *v1.Namespace
	switch t := obj.(type) {
	case *v1.Namespace:
		ns = t
	case cache.DeletedFinalStateUnknown:
		ns = convert2Namespace(obj)
	default:
		log.Logger().Warn("unable to convert to Namespace")
		return
	}
	if ns == nil {
		return
	}

	h.cache.Lock()
	defer h.cache.Unlock()
	delete(h.cache.nameSpaces, ns.Name)
}

// getAnnotationValues retrieves the annotation from the namespace.
// Converts the presence and content into a tri-state nsFlags object containing all nsFlags.
func getAnnotationValues(ns *v1.Namespace) nsFlags {
	if ns == nil {
		return nsFlags{-1, -1}
	}

	return nsFlags{
		enableYuniKorn: getAnnotationValue(ns.Annotations, constants.AnnotationEnableYuniKorn),
		generateAppID:  getAnnotationValue(ns.Annotations, constants.AnnotationGenerateAppID),
	}
}

// getAnnotationValue retrieves the value of name from the map and convert it to a tri-state.
// Returns -1 if name is not present, 1 if the value is set to "true", 0 for all other cases.
func getAnnotationValue(m map[string]string, name string) int {
	strVal, ok := m[name]
	if !ok {
		return -1
	}
	switch strVal {
	case constants.True:
		return 1
	default:
		return 0
	}
}
