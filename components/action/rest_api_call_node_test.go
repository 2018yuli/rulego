/*
 * Copyright 2023 The RuleGo Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package action

import (
	"github.com/2018yuli/rulego/api/types"
	"github.com/2018yuli/rulego/test"
	"github.com/2018yuli/rulego/test/assert"
	"testing"
)

func TestRestApiCallNodeOnMsg(t *testing.T) {
	var node RestApiCallNode
	var configuration = make(types.Configuration)
	configuration["restEndpointUrlPattern"] = "https://gitee.com"
	configuration["requestMethod"] = "POST"
	config := types.NewConfig()
	err := node.Init(config, configuration)
	if err != nil {
		t.Errorf("err=%s", err)
	}
	ctx := test.NewRuleContext(config, func(msg types.RuleMsg, relationType string) {
		code := msg.Metadata.GetValue(statusCode)
		assert.Equal(t, "404", code)
	})
	metaData := types.BuildMetadata(make(map[string]interface{}))
	msg := ctx.NewMsg("TEST_MSG_TYPE_AA", metaData, "{\"test\":\"AA\"}")
	err = node.OnMsg(ctx, msg)
	if err != nil {
		t.Errorf("err=%s", err)
	}

}
