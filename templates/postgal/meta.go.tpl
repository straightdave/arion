// Postgal
// auto-generated source code: meta.go
// by arion (https://github.com/straightdave/arion)
package main

import (
    "reflect"

    "github.com/golang/protobuf/proto"
)

const _lesDump = "{{ .LesDump }}"

func init() {
    _version  = "{{ .GeneratedTime }}"
}

func getVarByTypeName(typeName string) proto.Message {
    switch typeName {
    {{ range $element := .List -}}
    case "{{- $element -}}":
        return reflect.New(reflect.TypeOf(& {{- $element -}} {}).Elem()).Interface().(proto.Message)
    {{ end }}
    }
    return nil
}
