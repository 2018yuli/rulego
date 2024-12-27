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

package str

import (
	"fmt"
	"github.com/2018yuli/rulego/utils/json"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const varPatternLeft = "${"
const varPatternRight = "}"

func init() {
	//设置随机种子
	rand.Seed(time.Now().UnixNano())
}

// SprintfDict formats a string according to a pattern and a dictionary of variables.
// The pattern is a string that contains placeholders for the variables in the form of ${key}.
// The dict is a map from keys to values that will replace the placeholders.
// For example, SprintfDict(“Hello, ${name}!”, map[string]string{“name”: “Alice”}) returns “Hello, Alice!”.
// If the pattern contains a key that is not in the dict, it will be left unchanged.
// If the dict contains a key that is not in the pattern, it will be ignored.
func SprintfDict(pattern string, dict map[string]interface{}) string {
	var result = pattern
	for key, value := range dict {
		result = ProcessVar(result, key, value)
	}
	return result
}

func ProcessVar(pattern, key string, val interface{}) string {
	varPattern := varPatternLeft + key + varPatternRight
	return strings.Replace(pattern, varPattern, fmt.Sprintf("%s", val), -1)
}

const randomStrOptions = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const randomStrOptionsLen = len(randomStrOptions)

// RandomStr 创建指定长度的随机字符
func RandomStr(num int) string {
	var builder strings.Builder
	for i := 0; i < num; i++ {
		builder.WriteByte(randomStrOptions[rand.Intn(randomStrOptionsLen)])
	}
	return builder.String()
}

// ToString input的值转成字符串
func ToString(input interface{}) string {
	if input == nil {
		return ""
	}
	switch v := input.(type) {
	case string:
		return v
	case bool:
		return strconv.FormatBool(v)
	case float64:
		ft := input.(float64)
		return strconv.FormatFloat(ft, 'f', -1, 64)
	case float32:
		ft := input.(float32)
		return strconv.FormatFloat(float64(ft), 'f', -1, 32)
	case int:
		return strconv.Itoa(v)
	case uint:
		return strconv.Itoa(int(v))
	case int8:
		return strconv.Itoa(int(v))
	case uint8:
		return strconv.Itoa(int(v))
	case int16:
		return strconv.Itoa(int(v))
	case uint16:
		return strconv.Itoa(int(v))
	case int32:
		return strconv.Itoa(int(v))
	case uint32:
		return strconv.Itoa(int(v))
	case int64:
		return strconv.FormatInt(v, 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case []byte:
		return string(v)
	case fmt.Stringer:
		return v.String()
	case error:
		return v.Error()
	default:
		if newValue, err := json.Marshal(input); err == nil {
			return string(newValue)
		} else {
			return ""
		}
	}
}

// ToStringMapString 把interface类型 转 map[string]string类型
func ToStringMapString(input interface{}) map[string]interface{} {
	var output = map[string]interface{}{}

	switch v := input.(type) {
	case map[string]string:
		for k, val := range v {
			output[ToString(k)] = ToString(val)
		}
		return output
	case map[string]interface{}:
		for k, val := range v {
			output[ToString(k)] = ToString(val)
		}
		return output
	case map[interface{}]string:
		for k, val := range v {
			output[ToString(k)] = ToString(val)
		}
		return output
	case map[interface{}]interface{}:
		for k, val := range v {
			output[ToString(k)] = ToString(val)
		}
		return output
	case string:
		_ = json.Unmarshal([]byte(v), &output)
		return output
	default:
		return output
	}
}

// CheckHasVar 检查字符串是否有占位符
func CheckHasVar(str string) bool {
	return strings.Contains(str, "${") && strings.Contains(str, "}")
}

// ConvertDollarPlaceholder 转postgres风格占位符
func ConvertDollarPlaceholder(sql, dbType string) string {
	if dbType == "postgres" {
		n := 1
		for strings.Contains(sql, "?") {
			sql = strings.Replace(sql, "?", fmt.Sprintf("$%d", n), 1)
			n++
		}
	}
	return sql
}

// RemoveBraces A function that takes a string with ${} and returns a string without them
func RemoveBraces(s string) string {
	// Create a new empty string
	result := ""
	// Loop through each character in the input string
	for i := 0; i < len(s); i++ {
		// Get the current character
		c := s[i]
		// If the character is $, check the next character
		if c == '$' && i+1 < len(s) {
			// If the next character is {, skip it and move to the next one
			if s[i+1] == '{' {
				i++
				continue
			}
		}
		// If the character is }, skip it and move to the next one
		if c == '}' {
			continue
		}
		// If the character is a space, skip it and move to the next one
		if c == ' ' {
			continue
		}
		// Otherwise, append the character to the result string
		result += string(c)
	}
	// Return the result string
	return result
}
