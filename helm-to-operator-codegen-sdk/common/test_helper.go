/*
Copyright 2023 The Nephio Authors.

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
	"bufio"
	"io"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

type Tests struct {
	input    any
	expected any
}

/*
Reads the File from FilePath and returns the file-data
*/
func getFileContents(inputFilePath string) ([]byte, error) {
	file, err := os.Open(filepath.Clean(inputFilePath))
	if err != nil {
		return nil, err
	}
	fp := bufio.NewReader(file)
	data, err := io.ReadAll(fp)
	if err != nil {
		return nil, err
	}
	return data, nil

}

/*
Convert the KRM Resources to *unstructured.Unstructured (map[string]any)
Returns the *unstructured.Unstructured Object, GroupVersionKind (gvk), error
*/
func unstructuredDecode(data []byte) (*unstructured.Unstructured, *schema.GroupVersionKind, error) {
	obj := &unstructured.Unstructured{}
	// Decode YAML into unstructured.Unstructured
	dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, gvk, err := dec.Decode(data, nil, obj)
	if err != nil {
		return nil, nil, err
	}
	return obj, gvk, nil
}
