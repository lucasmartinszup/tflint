package loader

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/wata727/tflint/logger"
	"github.com/wata727/tflint/state"
)

func TestLoadHCL(t *testing.T) {
	cases := []struct {
		Name  string
		Input string
		Error bool
	}{
		{
			Name:  "return parsed object",
			Input: "template.tf",
			Error: false,
		},
		{
			Name:  "file not found",
			Input: "not_found.tf",
			Error: true,
		},
		{
			Name:  "invalid syntax file",
			Input: "invalid.tf",
			Error: true,
		},
	}

	for _, tc := range cases {
		prev, _ := filepath.Abs(".")
		dir, _ := os.Getwd()
		defer os.Chdir(prev)
		testDir := dir + "/test-fixtures/files"
		os.Chdir(testDir)

		_, err := loadHCL(tc.Input, logger.Init(false))
		if tc.Error == true && err == nil {
			t.Fatalf("should be happen error.\n\ntestcase: %s", tc.Name)
		}
		if tc.Error == false && err != nil {
			t.Fatalf("should not be happen error.\nError: %s\n\ntestcase: %s", err, tc.Name)
		}
	}
}

func TestLoadTemplate(t *testing.T) {
	type Input struct {
		ListMap map[string]*ast.ObjectList
		File    string
	}

	cases := []struct {
		Name   string
		Input  Input
		Result map[string]*ast.ObjectList
	}{
		{
			Name: "add file list",
			Input: Input{
				ListMap: map[string]*ast.ObjectList{
					"example.tf": &ast.ObjectList{},
				},
				File: "empty.tf",
			},
			Result: map[string]*ast.ObjectList{
				"example.tf": &ast.ObjectList{},
				"empty.tf":   &ast.ObjectList{},
			},
		},
	}

	for _, tc := range cases {
		prev, _ := filepath.Abs(".")
		dir, _ := os.Getwd()
		defer os.Chdir(prev)
		testDir := dir + "/test-fixtures/files"
		os.Chdir(testDir)
		load := &Loader{
			Logger:  logger.Init(false),
			ListMap: tc.Input.ListMap,
		}

		load.LoadTemplate(tc.Input.File)
		if !reflect.DeepEqual(load.ListMap, tc.Result) {
			t.Fatalf("Bad: %s\nExpected: %s\n\ntestcase: %s", load.ListMap, tc.Result, tc.Name)
		}
	}
}

func TestLoadModuleFile(t *testing.T) {
	type Input struct {
		Key string
		Src string
	}

	cases := []struct {
		Name   string
		Input  Input
		Result map[string]*ast.ObjectList
		Error  bool
	}{
		{
			Name: "load module",
			Input: Input{
				Key: "example",
				Src: "github.com/wata727/example-module",
			},
			Result: map[string]*ast.ObjectList{
				"github.com/wata727/example-module/main.tf":   &ast.ObjectList{},
				"github.com/wata727/example-module/output.tf": &ast.ObjectList{},
			},
			Error: false,
		},
		{
			Name: "module not found",
			Input: Input{
				Key: "not_found",
				Src: "github.com/wata727/example-module",
			},
			Result: make(map[string]*ast.ObjectList),
			Error:  true,
		},
	}

	for _, tc := range cases {
		prev, _ := filepath.Abs(".")
		dir, _ := os.Getwd()
		defer os.Chdir(prev)
		testDir := dir + "/test-fixtures/modules"
		os.Chdir(testDir)
		load := NewLoader(false)

		err := load.LoadModuleFile(tc.Input.Key, tc.Input.Src)
		if tc.Error == true && err == nil {
			t.Fatalf("should be happen error.\n\ntestcase: %s", tc.Name)
			continue
		}
		if tc.Error == false && err != nil {
			t.Fatalf("should not be happen error.\nError: %s\n\ntestcase: %s", err, tc.Name)
			continue
		}

		if !reflect.DeepEqual(load.ListMap, tc.Result) {
			t.Fatalf("Bad: %s\nExpected: %s\n\ntestcase: %s", load.ListMap, tc.Result, tc.Name)
		}
	}
}

func TestLoadAllTemplate(t *testing.T) {
	cases := []struct {
		Name   string
		Input  string
		Result map[string]*ast.ObjectList
		Error  bool
	}{
		{
			Name:  "load all files",
			Input: "all-files",
			Result: map[string]*ast.ObjectList{
				"all-files/main.tf":   &ast.ObjectList{},
				"all-files/output.tf": &ast.ObjectList{},
			},
			Error: false,
		},
		{
			Name:   "dir not found",
			Input:  "not_found",
			Result: make(map[string]*ast.ObjectList),
			Error:  true,
		},
	}

	for _, tc := range cases {
		prev, _ := filepath.Abs(".")
		dir, _ := os.Getwd()
		defer os.Chdir(prev)
		testDir := dir + "/test-fixtures"
		os.Chdir(testDir)
		load := NewLoader(false)

		err := load.LoadAllTemplate(tc.Input)
		if tc.Error == true && err == nil {
			t.Fatalf("should be happen error.\n\ntestcase: %s", tc.Name)
			continue
		}
		if tc.Error == false && err != nil {
			t.Fatalf("should not be happen error.\nError: %s\n\ntestcase: %s", err, tc.Name)
			continue
		}

		if !reflect.DeepEqual(load.ListMap, tc.Result) {
			t.Fatalf("Bad: %s\nExpected: %s\n\ntestcase: %s", load.ListMap, tc.Result, tc.Name)
		}
	}
}

func TestLoadState(t *testing.T) {
	cases := []struct {
		Name   string
		Dir    string
		Result *state.TFState
	}{
		{
			Name: "load local state",
			Dir:  "local-state",
			Result: &state.TFState{
				Modules: []*state.Module{
					&state.Module{
						Resources: map[string]*state.Resource{
							"aws_db_parameter_group.production": &state.Resource{
								Type:         "aws_db_parameter_group",
								Dependencies: []string{},
								Primary: &state.Instance{
									ID: "production",
									Attributes: map[string]string{
										"arn":         "arn:aws:rds:us-east-1:hogehoge:pg:production",
										"description": "production-db-parameter-group",
										"family":      "mysql5.6",
										"id":          "production",
										"name":        "production",
										"parameter.#": "0",
										"tags.%":      "0",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			Name: "load remote state",
			Dir:  "remote-state",
			Result: &state.TFState{
				Modules: []*state.Module{
					&state.Module{
						Resources: map[string]*state.Resource{
							"aws_db_parameter_group.staging": &state.Resource{
								Type:         "aws_db_parameter_group",
								Dependencies: []string{},
								Primary: &state.Instance{
									ID: "staging",
									Attributes: map[string]string{
										"arn":         "arn:aws:rds:us-east-1:hogehoge:pg:staging",
										"description": "staging-db-parameter-group",
										"family":      "mysql5.6",
										"id":          "staging",
										"name":        "staging",
										"parameter.#": "0",
										"tags.%":      "0",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			Name:   "state not found",
			Dir:    "files",
			Result: &state.TFState{},
		},
	}

	for _, tc := range cases {
		prev, _ := filepath.Abs(".")
		dir, _ := os.Getwd()
		defer os.Chdir(prev)
		testDir := dir + "/test-fixtures/" + tc.Dir
		os.Chdir(testDir)
		load := NewLoader(false)

		load.LoadState()
		if !reflect.DeepEqual(load.State, tc.Result) {
			t.Fatalf("Bad: %s\nExpected: %s\n\ntestcase: %s", load.State, tc.Result, tc.Name)
		}
		os.Chdir(prev)
	}
}

func TestDump(t *testing.T) {
	load := NewLoader(false)
	listMap := map[string]*ast.ObjectList{
		"main.tf":   &ast.ObjectList{},
		"output.tf": &ast.ObjectList{},
	}
	state := &state.TFState{
		Modules: []*state.Module{
			&state.Module{
				Resources: map[string]*state.Resource{
					"aws_db_parameter_group.production": &state.Resource{
						Type:         "aws_db_parameter_group",
						Dependencies: []string{},
						Primary: &state.Instance{
							ID: "production",
							Attributes: map[string]string{
								"arn":         "arn:aws:rds:us-east-1:hogehoge:pg:production",
								"description": "production-db-parameter-group",
								"family":      "mysql5.6",
								"id":          "production",
								"name":        "production",
								"parameter.#": "0",
								"tags.%":      "0",
							},
						},
					},
				},
			},
		},
	}
	load.ListMap = listMap
	load.State = state

	dumpListMap, dumpState := load.Dump()
	if !reflect.DeepEqual(dumpListMap, listMap) {
		t.Fatalf("Bad: %s\nExpected: %s\n\n", dumpListMap, listMap)
	}
	if !reflect.DeepEqual(dumpState, state) {
		t.Fatalf("Bad: %s\nExpected: %s\n\n", dumpState, state)
	}
}