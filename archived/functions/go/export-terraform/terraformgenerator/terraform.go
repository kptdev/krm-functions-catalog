// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package terraformgenerator

import (
	"fmt"
	"math"
	"strings"
	"text/template"
	"time"
)

func (rs *terraformResources) getHCL() (map[string]string, error) {
	tmplUtilFns := template.FuncMap{
		"msToDays":              msToDays,
		"sToDays":               func(t int) (float64, error) { return msToDays(t * 1000) },
		"strSliceToCommaSepStr": func(s []string) string { return strings.Join(s, ",") },
	}

	tmpl, err := template.New("").Funcs(tmplUtilFns).ParseFS(templates, "templates/*")
	if err != nil {
		return nil, err
	}

	groupedResources := rs.getGrouped()

	data := make(map[string]string)
	resourceFiles := []string{"folders.tf", "iam.tf", "projects.tf", "log-export.tf", "network.tf"}
	for _, file := range resourceFiles {
		err := addFile(tmpl, file, groupedResources, data)
		if err != nil {
			return nil, err
		}
	}

	// only add other files if resource files exist
	metaFiles := []string{"README.md", "versions.tf", "variables.tf"}
	if len(data) > 0 {
		for _, file := range metaFiles {
			err := addFile(tmpl, file, rs, data)
			if err != nil {
				return nil, err
			}
		}
	}

	return data, nil
}

func addFile(tmpl *template.Template, name string, inputData interface{}, data map[string]string) error {
	builder := strings.Builder{}
	wr := &(builder)

	err := tmpl.ExecuteTemplate(wr, name, inputData)
	if err != nil {
		return err
	}

	content := strings.TrimSpace(builder.String())

	if len(content) < 1 {
		return nil
	}

	content = content + "\n"
	data[name] = content

	return nil
}

// msToDays converts milliseconds to days rounded to two decimal places
func msToDays(t int) (float64, error) {
	d, err := time.ParseDuration(fmt.Sprintf("%dms", t))
	if err != nil {
		return 0, err
	}
	return math.Round((d.Hours()/24)*100) / 100, nil
}
