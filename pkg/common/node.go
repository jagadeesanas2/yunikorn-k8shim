/*
Copyright 2019 The Unity Scheduler Authors

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

package common

import (
	"github.infra.cloudera.com/yunikorn/scheduler-interface/lib/go/si"
	"k8s.io/api/core/v1"
)

// stores info about what scheduler cares about a node
type Node struct {
	name string
	uid string
	resource *si.Resource
}

func CreateFrom(node *v1.Node) Node {
	return Node{
		name: node.Name,
		uid: string(node.UID),
		resource: GetNodeResource(&node.Status),
	}
}

func CreateFromNodeSpec(nodeName string, nodeUid string, nodeResource *si.Resource) Node {
	return Node {
		name: nodeName,
		uid: nodeUid,
		resource: nodeResource,
	}
}