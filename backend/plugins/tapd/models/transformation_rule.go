/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package models

import (
	"encoding/json"
	"github.com/apache/incubator-devlake/core/models/common"
)

type TapdTransformationRule struct {
	common.Model               `mapstructure:"-"`
	ConnectionId               uint64          `mapstructure:"connectionId" json:"connectionId"`
	Name                       string          `gorm:"type:varchar(255);index:idx_name_tapd,unique" validate:"required" mapstructure:"name" json:"name"`
	RemotelinkCommitShaPattern string          `mapstructure:"remotelinkCommitShaPattern,omitempty" json:"remotelinkCommitShaPattern" gorm:"type:varchar(255)"`
	RemotelinkRepoPattern      json.RawMessage `mapstructure:"remotelinkRepoPattern,omitempty" json:"remotelinkRepoPattern"`
	TypeMappings               json.RawMessage `mapstructure:"typeMappings,omitempty" json:"typeMappings"`
	StatusMappings             json.RawMessage `mapstructure:"statusMappings,omitempty" json:"statusMappings"`
}

func (t TapdTransformationRule) TableName() string {
	return "_tool_tapd_transformation_rules"
}