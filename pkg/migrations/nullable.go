// SPDX-License-Identifier: Apache-2.0

package migrations

import (
	"encoding/json"
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

type Nullable[T any] map[bool]T

func NewNullableWithValue[T any](value T) Nullable[T] {
	return Nullable[T]{true: value}
}

func NewNullNullable[T any]() Nullable[T] {
	var n Nullable[T]
	n.SetNull()
	return n
}

func (n Nullable[T]) IsNull() bool {
	_, ok := n[false]
	return !ok
}

func (t *Nullable[T]) SetNull() {
	var empty T
	*t = map[bool]T{false: empty}
}

func (n Nullable[T]) Get() (T, error) {
	var value T
	if n.IsNull() {
		return value, errors.New("nullable is null")
	}
	if !n.IsSpecified() {
		return value, errors.New("nullable is not specified")
	}
	return n[true], nil
}

func (n Nullable[T]) MustGet() T {
	value, err := n.Get()
	if err != nil {
		panic(err)
	}
	return value
}

func (n Nullable[T]) IsSpecified() bool {
	return len(n) != 0
}

func (n Nullable[T]) MarshalJSON() ([]byte, error) {
	if n.IsNull() {
		return []byte("null"), nil
	}
	return json.Marshal(n[true])
}

func (n *Nullable[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.SetNull()
		return nil
	}
	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*n = NewNullableWithValue(value)
	return nil
}

func (n *Nullable[T]) UnmarshalYAML(node *yaml.Node) error {
	fmt.Println(node)
	var value T
	if err := node.Decode(&value); err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(value)
	*n = NewNullableWithValue(value)
	return nil
}
