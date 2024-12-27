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

package types

import (
	"github.com/rulego/rulego/pool"
	"math"
	"time"
)

// Config 规则引擎配置
type Config struct {
	//OnDebug 节点调试信息回调函数，只有节点debugMode=true才会调用
	OnDebug func(flowType string, nodeId string, msg RuleMsg, relationType string, err error)
	//OnEnd 规则链执行完成回调函数，如果有多个结束点，则执行多次
	OnEnd func(msg RuleMsg, err error)
	//JsMaxExecutionTime js脚本执行超时时间，默认2000毫秒
	JsMaxExecutionTime time.Duration
	//Pool 协程池接口
	//如果不配置，则使用 go func 方式
	//默认使用`pool.WorkerPool`。兼容ants协程池，可以使用ants协程池实现
	//例如：
	//	pool, _ := ants.NewPool(math.MaxInt32)
	//	config := rulego.NewConfig(types.WithPool(pool))
	Pool Pool
	//ComponentsRegistry 组件库
	//默认使用`rulego.Registry`
	ComponentsRegistry ComponentRegistry
	//规则链解析接口，默认使用：`rulego.JsonParser`
	Parser Parser
	//Logger 日志记录接口，默认使用：`DefaultLogger()`
	Logger Logger
}

// Option is a function type that modifies the Config.
type Option func(*Config) error

func NewConfig(opts ...Option) Config {
	// Create a new Config with default values.
	c := &Config{
		JsMaxExecutionTime: time.Millisecond * 2000,
		Logger:             DefaultLogger(),
	}

	// Apply the options to the Config.
	for _, opt := range opts {
		_ = opt(c)
	}
	return *c
}

func DefaultPool() Pool {
	wp := &pool.WorkerPool{MaxWorkersCount: math.MaxInt32}
	wp.Start()
	return wp
}

// WithComponentsRegistry is an option that sets the components registry of the Config.
func WithComponentsRegistry(componentsRegistry ComponentRegistry) Option {
	return func(c *Config) error {
		c.ComponentsRegistry = componentsRegistry
		return nil
	}
}

// WithOnDebug is an option that sets the on debug callback of the Config.
func WithOnDebug(onDebug func(flowType string, nodeId string, msg RuleMsg, relationType string, err error)) Option {
	return func(c *Config) error {
		c.OnDebug = onDebug
		return nil
	}
}

// WithOnEnd is an option that sets the on end callback of the Config.
func WithOnEnd(onEnd func(msg RuleMsg, err error)) Option {
	return func(c *Config) error {
		c.OnEnd = onEnd
		return nil
	}
}

// WithPool is an option that sets the pool of the Config.
func WithPool(pool Pool) Option {
	return func(c *Config) error {
		c.Pool = pool
		return nil
	}
}

func WithDefaultPool() Option {
	return func(c *Config) error {
		wp := &pool.WorkerPool{MaxWorkersCount: math.MaxInt32}
		wp.Start()
		c.Pool = wp
		return nil
	}
}

// WithJsMaxExecutionTime is an option that sets the js max execution time of the Config.
func WithJsMaxExecutionTime(jsMaxExecutionTime time.Duration) Option {
	return func(c *Config) error {
		c.JsMaxExecutionTime = jsMaxExecutionTime
		return nil
	}
}

// WithParser is an option that sets the parser of the Config.
func WithParser(parser Parser) Option {
	return func(c *Config) error {
		c.Parser = parser
		return nil
	}
}

// WithLogger is an option that sets the logger of the Config.
func WithLogger(logger Logger) Option {
	return func(c *Config) error {
		c.Logger = logger
		return nil
	}
}
