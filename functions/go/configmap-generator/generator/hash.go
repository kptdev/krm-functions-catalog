package generator

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
)

func Hash(object *fn.KubeObject) (string, error) {
	content, err := effectiveContent(object)
	if err != nil {
		return "", nil
	}
	hex256 := fmt.Sprintf("%x", sha256.Sum256([]byte(content)))
	return encode(hex256)
}

// Copied from https://github.com/kubernetes/kubernetes
// /blob/master/pkg/kubectl/util/hash/hash.go
func encode(hex string) (string, error) {
	if len(hex) < 10 {
		return "", fmt.Errorf(
			"input length must be at least 10")
	}
	enc := []rune(hex[:10])
	for i := range enc {
		switch enc[i] {
		case '0':
			enc[i] = 'g'
		case '1':
			enc[i] = 'h'
		case '3':
			enc[i] = 'k'
		case 'a':
			enc[i] = 'm'
		case 'e':
			enc[i] = 't'
		}
	}
	return string(enc), nil
}

func effectiveContent(object *fn.KubeObject) (string, error) {
	data, _, err := object.NestedStringMap("data")
	if err != nil {
		return "", err
	}
	m := map[string]interface{}{
		"kind": object.GetKind(),
		"name": object.GetName(),
		"data": data,
	}
	// skip binaryData
	content, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
