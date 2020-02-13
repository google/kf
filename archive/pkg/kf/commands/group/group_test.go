// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package group_test

import (
	"bytes"
	"testing"

	"github.com/google/kf/pkg/kf/commands/group"
	"github.com/google/kf/pkg/kf/testutil"
	"github.com/spf13/cobra"
)

func TestCalculateMinWidth(t *testing.T) {
	cases := map[string]struct {
		groups        group.CommandGroups
		expectedWidth int
	}{
		"zero-lenth array": {
			groups:        group.CommandGroups{},
			expectedWidth: 0,
		},
		"one empty group": {
			groups: group.CommandGroups{
				{
					Name:     "group-1",
					Commands: []*cobra.Command{},
				},
			},
			expectedWidth: 0,
		},
		"group with nil": {
			groups: group.CommandGroups{
				{
					Name: "group-1",
					Commands: []*cobra.Command{
						nil,
					},
				},
			},
			expectedWidth: 0,
		},
		"group with command": {
			groups: group.CommandGroups{
				{
					Name: "group-1",
					Commands: []*cobra.Command{
						{
							Use: "command",
						},
					},
				},
			},
			expectedWidth: len("command"),
		},
		"group with a few commands": {
			groups: group.CommandGroups{
				{
					Name: "group-1",
					Commands: []*cobra.Command{
						{
							Use: "command",
						},
						{
							Use: "foo",
						},
						{
							Use: "foobarcommand",
						},
					},
				},
			},
			expectedWidth: len("foobarcommand"),
		},
		"a few groups with a few commands": {
			groups: group.CommandGroups{
				{
					Name: "group-1",
					Commands: []*cobra.Command{
						{
							Use: "command",
						},
						{
							Use: "foo",
						},
						{
							Use: "foobarcommand",
						},
					},
				},
				{
					Name: "group-2",
					Commands: []*cobra.Command{
						{
							Use: "command2",
						},
						{
							Use: "reallylongcommandnooneuses",
						},
						{
							Use: "1234 spaces in here",
						},
					},
				},
				{
					Name: "group-3",
					Commands: []*cobra.Command{
						{
							Use: "some-command",
						},
						{
							Use: "another-command",
						},
						{
							Use: "get-new-command-for-use",
						},
					},
				},
			},
			expectedWidth: len("reallylongcommandnooneuses"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			w := tc.groups.CalculateMinWidth()
			if w != tc.expectedWidth {
				t.Errorf("Expected minWidth to be %d actual value %d", tc.expectedWidth, w)
			}
		})
	}
}

func TestPrintTrimmedMultilineString(t *testing.T) {
	cases := map[string]struct {
		str      string
		expected string
	}{
		"empty string": {
			str:      "",
			expected: "",
		},
		"one string": {
			str:      "some-text",
			expected: "some-text\n",
		},
		"two strings": {
			str:      "some-text\nand more",
			expected: "some-text\nand more\n",
		},
		"trim": {
			str:      "   some-text \n      and more    ",
			expected: "some-text\nand more\n",
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			var b bytes.Buffer
			group.PrintTrimmedMultilineString(tc.str, &b)

			actual := b.String()
			testutil.AssertEqual(t, "output", tc.expected, actual)
		})
	}
}
