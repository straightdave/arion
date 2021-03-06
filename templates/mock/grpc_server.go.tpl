//
// This file is generated by Arion.
//

package main

import (
	"context"
	"encoding/json"
	"log"
)

type server struct {
	{{- .PBServerInterface -}}
}

{{ range $method := .Methods }}
func (s *server) {{ $method.Name -}}(ctx context.Context, in *{{- $method.RequestType -}}) (*{{- $method.ResponseType -}}, error) {
	respMapGuard.RLock()
	rawjson := respMap["{{- $method.Name -}}"]
	respMapGuard.RUnlock()

	respData := &{{- $method.ResponseType -}}{}

	// eat error
	_ = json.Unmarshal(rawjson, respData)

	log.Printf("Calling {{ $method.Name }}, respond: %v", respData)
	return respData, nil
}

{{ end }}
