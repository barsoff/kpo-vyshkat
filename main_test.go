package main

import (
	"bytes"
	"strings"
	"testing"

	"hw1/internal/cli"
	"hw1/internal/fleet"
	"hw1/internal/model"
)

func TestRunModes(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		input   string
		wantErr bool
		output  string
	}{
		{"interactive", nil, "report\nexit\n", false, "Транспорт: 0"},
		{"demo", []string{"-demo"}, "", false, "Суточный расход энергии: 7.00"},
		{"help", []string{"-help"}, "", false, "-demo"},
		{"bad_flag", []string{"-unknown"}, "", true, ""},
		{"extra_argument", []string{"unexpected"}, "", true, ""},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			err := run(test.args, strings.NewReader(test.input), &output)
			if (err != nil) != test.wantErr {
				t.Fatalf("run error: %v", err)
			}
			if !strings.Contains(output.String(), test.output) {
				t.Fatalf("missing %q in output:\n%s", test.output, output.String())
			}
		})
	}
}

type rejectingInspector struct{}

func (rejectingInspector) Inspect(model.Transport) bool {
	return false
}

func TestDigInjectsInspectorIntoApplication(t *testing.T) {
	input := strings.NewReader("add transport\nbicycle\nBike\nT-1\n6\nнет\nadd item\nhelmet\nHelmet\nI-1\nexit\n")
	var output bytes.Buffer
	container, err := buildContainer(input, &output, func() fleet.Inspector {
		return rejectingInspector{}
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := container.Invoke(func(console *cli.Console) error { return console.Run() }); err != nil {
		t.Fatal(err)
	}
	if err := container.Invoke(func(park *fleet.Fleet) {
		if park.Summary() != (fleet.Summary{ThingCount: 1, TotalItems: 1}) {
			t.Fatal("console and report received different fleet instances, or fake was ignored")
		}
	}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "техосмотр отклонил") {
		t.Fatalf("DI replacement did not control admission: %s", output.String())
	}
}
