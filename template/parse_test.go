//go:build !windows
// +build !windows

package template

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func boolPointer(tf bool) *bool {
	return &tf
}

func TestParse(t *testing.T) {
	cases := []struct {
		File   string
		Result *Template
		Err    bool
	}{
		/*
		 * Builders
		 */
		{
			"parse-basic.json",
			&Template{
				Builders: map[string]*Builder{
					"something": {
						Name: "something",
						Type: "something",
					},
				},
			},
			false,
		},
		{
			"parse-basic-config.json",
			&Template{
				Builders: map[string]*Builder{
					"something": {
						Name: "something",
						Type: "something",
						Config: map[string]interface{}{
							"foo": "bar",
						},
					},
				},
			},
			false,
		},
		{
			"parse-builder-no-type.json",
			nil,
			true,
		},
		{
			"parse-builder-repeat.json",
			nil,
			true,
		},

		/*
		 * Provisioners
		 */
		{
			"parse-provisioner-basic.json",
			&Template{
				Provisioners: []*Provisioner{
					{
						Type: "something",
					},
				},
			},
			false,
		},
		{
			"parse-provisioner-config.json",
			&Template{
				Provisioners: []*Provisioner{
					{
						Type: "something",
						Config: map[string]interface{}{
							"inline": "echo 'foo'",
						},
					},
				},
			},
			false,
		},
		{
			"parse-provisioner-pause-before.json",
			&Template{
				Provisioners: []*Provisioner{
					{
						Type:        "something",
						PauseBefore: 1 * time.Second,
					},
				},
			},
			false,
		},

		{
			"parse-provisioner-retry.json",
			&Template{
				Provisioners: []*Provisioner{
					{
						Type:       "something",
						MaxRetries: "5",
					},
				},
			},
			false,
		},

		{
			"parse-provisioner-timeout.json",
			&Template{
				Provisioners: []*Provisioner{
					{
						Type:    "something",
						Timeout: 5 * time.Minute,
					},
				},
			},
			false,
		},

		{
			"parse-provisioner-only.json",
			&Template{
				Provisioners: []*Provisioner{
					{
						Type: "something",
						OnlyExcept: OnlyExcept{
							Only: []string{"foo"},
						},
					},
				},
			},
			false,
		},

		{
			"parse-provisioner-except.json",
			&Template{
				Provisioners: []*Provisioner{
					{
						Type: "something",
						OnlyExcept: OnlyExcept{
							Except: []string{"foo"},
						},
					},
				},
			},
			false,
		},

		{
			"parse-provisioner-override.json",
			&Template{
				Provisioners: []*Provisioner{
					{
						Type: "something",
						Override: map[string]interface{}{
							"foo": map[string]interface{}{},
						},
					},
				},
			},
			false,
		},

		{
			"parse-provisioner-no-type.json",
			nil,
			true,
		},

		{
			"parse-variable-default.json",
			&Template{
				Variables: map[string]*Variable{
					"foo": {
						Default: "foo",
						Key:     "foo",
					},
				},
			},
			false,
		},

		{
			"parse-variable-required.json",
			&Template{
				Variables: map[string]*Variable{
					"foo": {
						Required: true,
						Key:      "foo",
					},
				},
			},
			false,
		},

		{
			"parse-pp-basic.json",
			&Template{
				PostProcessors: [][]*PostProcessor{
					{
						{
							Name: "foo",
							Type: "foo",
							Config: map[string]interface{}{
								"foo": "bar",
							},
						},
					},
				},
			},
			false,
		},

		{
			"parse-pp-keep.json",
			&Template{
				PostProcessors: [][]*PostProcessor{
					{
						{
							Name:              "foo",
							Type:              "foo",
							KeepInputArtifact: boolPointer(true),
						},
					},
				},
			},
			false,
		},

		{
			"parse-pp-only.json",
			&Template{
				PostProcessors: [][]*PostProcessor{
					{
						{
							Name: "foo",
							Type: "foo",
							OnlyExcept: OnlyExcept{
								Only: []string{"bar"},
							},
						},
					},
				},
			},
			false,
		},

		{
			"parse-pp-except.json",
			&Template{
				PostProcessors: [][]*PostProcessor{
					{
						{
							Name: "foo",
							Type: "foo",
							OnlyExcept: OnlyExcept{
								Except: []string{"bar"},
							},
						},
					},
				},
			},
			false,
		},

		{
			"parse-pp-string.json",
			&Template{
				PostProcessors: [][]*PostProcessor{
					{
						{
							Name: "foo",
							Type: "foo",
						},
					},
				},
			},
			false,
		},

		{
			"parse-pp-map.json",
			&Template{
				PostProcessors: [][]*PostProcessor{
					{
						{
							Name: "foo",
							Type: "foo",
						},
					},
				},
			},
			false,
		},

		{
			"parse-pp-slice.json",
			&Template{
				PostProcessors: [][]*PostProcessor{
					{
						{
							Name: "foo",
							Type: "foo",
						},
					},
					{
						{
							Name: "bar",
							Type: "bar",
						},
					},
				},
			},
			false,
		},

		{
			"parse-pp-multi.json",
			&Template{
				PostProcessors: [][]*PostProcessor{
					{
						{
							Name: "foo",
							Type: "foo",
						},
						{
							Name: "bar",
							Type: "bar",
						},
					},
				},
			},
			false,
		},

		{
			"parse-pp-no-type.json",
			nil,
			true,
		},

		{
			"parse-description.json",
			&Template{
				Description: "foo",
			},
			false,
		},

		{
			"parse-min-version.json",
			&Template{
				MinVersion: "1.2",
			},
			false,
		},

		{
			"parse-comment.json",
			&Template{
				Builders: map[string]*Builder{
					"something": {
						Name: "something",
						Type: "something",
					},
				},
				Comments: map[string]string{
					"_info": "foo",
				},
			},
			false,
		},
		{
			"parse-monolithic.json",
			&Template{
				Comments: map[string]string{
					"_comment": "comment",
				},
				Description: "Description Test",
				MinVersion:  "1.3.0",
				SensitiveVariables: []*Variable{
					{
						Required: false,
						Key:      "one",
						Default:  "1",
					},
				},
				Variables: map[string]*Variable{
					"one": {
						Required: false,
						Key:      "one",
						Default:  "1",
					},
					"two": {
						Required: false,
						Key:      "two",
						Default:  "2",
					},
					"three": {
						Required: true,
						Key:      "three",
						Default:  "",
					},
				},
				Builders: map[string]*Builder{
					"amazon-ebs": {
						Name: "amazon-ebs",
						Type: "amazon-ebs",
						Config: map[string]interface{}{
							"ami_name":      "AMI Name",
							"instance_type": "t2.micro",
							"ssh_username":  "ec2-user",
							"source_ami":    "ami-aaaaaaaaaaaaaa",
						},
					},
					"docker": {
						Name: "docker",
						Type: "docker",
						Config: map[string]interface{}{
							"image":       "ubuntu",
							"export_path": "image.tar",
						},
					},
				},
				Provisioners: []*Provisioner{
					{
						Type: "shell",
						Config: map[string]interface{}{
							"script": "script.sh",
						},
					},
					{
						Type: "shell",
						Config: map[string]interface{}{
							"script": "script.sh",
						},
						Override: map[string]interface{}{
							"docker": map[string]interface{}{
								"execute_command": "echo 'override'",
							},
						},
					},
				},
				PostProcessors: [][]*PostProcessor{
					{
						{
							Name: "compress",
							Type: "compress",
						},
						{
							Name: "vagrant",
							Type: "vagrant",
							OnlyExcept: OnlyExcept{
								Only: []string{"docker"},
							},
						},
					},
					{
						{
							Name: "shell-local",
							Type: "shell-local",
							Config: map[string]interface{}{
								"inline": []interface{}{"echo foo"},
							},
							OnlyExcept: OnlyExcept{
								Except: []string{"amazon-ebs"},
							},
						},
					},
				},
			},
			false,
		},
	}

	for i, tc := range cases {
		path, _ := filepath.Abs(fixtureDir(tc.File))
		tpl, err := ParseFile(fixtureDir(tc.File))
		if (err != nil) != tc.Err {
			t.Fatalf("%s\n\nerr: %s", tc.File, err)
		}

		if tc.Result != nil {
			tc.Result.Path = path
		}
		if tpl != nil {
			tpl.RawContents = nil
		}
		if diff := cmp.Diff(tpl, tc.Result); diff != "" {
			t.Fatalf("[%d]bad: %s\n%v", i, tc.File, diff)
		}

		// Only test template writing if the template is valid
		if tc.Err == false {
			// Get rawTemplate
			raw, err := tpl.Raw()
			if err != nil {
				t.Fatalf("Failed to convert back to raw template: %s\n\n%v\n\n%s", tc.File, tpl, err)
			}

			out, _ := json.MarshalIndent(raw, "", "  ")
			if err != nil {
				t.Fatalf("Failed to marshal raw template: %s\n\n%v\n\n%s", tc.File, raw, err)
			}

			// Write JSON to a buffer (emulates filesystem write without dirtying the workspace)
			fileBuf := bytes.NewBuffer(out)

			// Parse the JSON template we wrote to our buffer
			tplRewritten, err := Parse(fileBuf)
			if err != nil {
				t.Fatalf("Failed to re-read marshalled template: %s\n\n%v\n\n%s\n\n%s", tc.File, tpl, out, err)
			}

			// Override the metadata we don't care about (file path, raw file contents)
			tplRewritten.Path = path
			tplRewritten.RawContents = nil

			// Test that our output raw template is functionally equal
			if !reflect.DeepEqual(tpl, tplRewritten) {
				t.Fatalf("Data lost when writing template to file: %s\n\n%v\n\n%v\n\n%s", tc.File, tpl, tplRewritten, out)
			}
		}
	}
}

func TestParse_contents(t *testing.T) {
	tpl, err := ParseFile(fixtureDir("parse-contents.json"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	actual := strings.TrimSpace(string(tpl.RawContents))
	expected := `{"builders":[{"type":"test"}]}`
	if actual != expected {
		t.Fatalf("bad: %s\n\n%s", actual, expected)
	}
}

func TestParse_bad(t *testing.T) {
	cases := []struct {
		File     string
		Expected string
	}{
		{"error-beginning.json", "line 1, column 1 (offset 1)"},
		{"error-middle.json", "line 5, column 6 (offset 50)"},
		{"error-end.json", "line 1, column 30 (offset 30)"},
		{"malformed.json", "line 16, column 3 (offset 433)"},
	}
	for _, tc := range cases {
		_, err := ParseFile(fixtureDir(tc.File))
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(err.Error(), tc.Expected) {
			t.Fatalf("file: %s\nExpected: %s\n%s\n", tc.File, tc.Expected, err.Error())
		}
	}
}

func TestParse_checkForDuplicateFields(t *testing.T) {
	cases := []struct {
		File     string
		Expected string
	}{
		{"error-duplicate-variables.json", "template has duplicate field: variables"},
		{"error-duplicate-config.json", "template has duplicate field: foo"},
	}
	for _, tc := range cases {
		_, err := ParseFile(fixtureDir(tc.File))
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(err.Error(), tc.Expected) {
			t.Fatalf("file: %s\nExpected: %s\n%s\n", tc.File, tc.Expected, err.Error())
		}
	}
}
