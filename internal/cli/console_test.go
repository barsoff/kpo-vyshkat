package cli_test

import (
	"bytes"
	"errors"
	"io"
	"math"
	"strings"
	"testing"

	"hw1/internal/cli"
	"hw1/internal/fleet"
)

func runScript(t *testing.T, script string) (*fleet.Fleet, string) {
	t.Helper()
	park := fleet.NewFleet(&fleet.ServiceCenter{})
	var output bytes.Buffer
	if err := cli.New(park, strings.NewReader(script), &output).Run(); err != nil {
		t.Fatal(err)
	}
	return park, output.String()
}

func TestEmptyReportsAndHelp(t *testing.T) {
	_, output := runScript(t, "help\nreport\nbeginners\ninventory\nunknown\n\nexit\n")
	for _, expected := range []string{"add transport", "Транспорт: 0; вещи: 0", "0.00 кВт·ч", "Подходящего транспорта нет", "Имущество отсутствует", "Неизвестная команда", "До встречи"} {
		if !strings.Contains(output, expected) {
			t.Errorf("missing %q in output:\n%s", expected, output)
		}
	}
}

func TestAllKindsCanBeAddedThroughConsole(t *testing.T) {
	script := ` ADD   TRANSPORT
bicycle
Велосипед Кампус
T-1
6
нет
add transport
scooter
Самокат
T-2
10
no
0,5
add transport
ebike
Электровелосипед
T-3
5
false
0.8
add transport
surf
Электросёрф
T-4
1
0
2.2
add item
helmet
Шлем
I-1
add item
charger
Шкаф
I-2
1.5
add item
vest
Жилет
I-3
add item
dryer
Сушилка
I-4
2
report
beginners
inventory
exit
`
	park, output := runScript(t, script)
	summary := park.Summary()
	if summary.TransportCount != 4 || summary.ThingCount != 4 || math.Abs(summary.DailyEnergy-7) > 1e-9 {
		t.Fatalf("unexpected summary: %+v\n%s", summary, output)
	}
	if got := park.ForBeginners(); len(got) != 2 || got[0].InventoryNumber() != "T-2" || got[1].InventoryNumber() != "T-1" {
		t.Fatalf("wrong beginner selection: %+v", got)
	}
	for _, expected := range []string{"Велосипед Кампус [T-1]", "Жилет [I-3]", "Сушилка [I-4]", "Суточный расход энергии: 7.00"} {
		if !strings.Contains(output, expected) {
			t.Errorf("missing %q", expected)
		}
	}
	if strings.Contains(output, "Ошибка:") {
		t.Fatal(output)
	}
}

func TestInputErrorsLeaveFleetUnchanged(t *testing.T) {
	cases := []struct {
		name   string
		script string
	}{
		{"unknown_transport", "add transport\nboat\n"},
		{"unknown_thing", "add item\nrocket\n"},
		{"not_integer", "add transport\nbicycle\nBike\nT-1\nsix\n"},
		{"bad_boolean", "add transport\nbicycle\nBike\nT-1\n6\nmaybe\n"},
		{"bad_energy", "add transport\nscooter\nScooter\nT-1\n6\nнет\nenergy\n"},
		{"negative_energy", "add item\ncharger\nCharger\nI-1\n-1\n"},
		{"nan_energy", "add item\ndryer\nDryer\nI-1\nNaN\n"},
		{"out_of_range", "add transport\nbicycle\nBike\nT-1\n11\nнет\n"},
		{"empty_name", "add item\nhelmet\n   \nI-1\n"},
		{"empty_number", "add item\nvest\nVest\n   \n"},
		{"damaged", "add transport\nsurf\nSurf\nT-1\n6\nда\n1\n"},
		{"damaged_yes", "add transport\nbicycle\nBike\nT-1\n6\nyes\n"},
		{"damaged_true", "add transport\nebike\nBike\nT-1\n6\ntrue\n1\n"},
		{"damaged_one", "add transport\nbicycle\nBike\nT-1\n6\n1\n"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			park, output := runScript(t, test.script+"report\nexit\n")
			if park.Summary() != (fleet.Summary{}) || !strings.Contains(output, "Ошибка:") || !strings.Contains(output, "До встречи") {
				t.Fatalf("invalid input must not change state or terminate CLI:\n%s", output)
			}
		})
	}
}

func TestDuplicateThroughConsole(t *testing.T) {
	park, output := runScript(t, "add item\nhelmet\nШлем\nSAME\nadd transport\nscooter\nСамокат\nSAME\n6\nнет\n10\nreport\nexit\n")
	if park.Summary() != (fleet.Summary{ThingCount: 1, TotalItems: 1}) || !strings.Contains(output, "инвентарный номер уже занят") {
		t.Fatalf("duplicate changed state:\n%s", output)
	}
}

func TestEndOfInputNeverAddsPartialObject(t *testing.T) {
	forms := [][]string{
		{"add transport", "bicycle", "Bike", "T-1", "6", "нет"},
		{"add transport", "scooter", "Scooter", "T-1", "6", "нет", "0.5"},
		{"add item", "helmet", "Helmet", "I-1"},
		{"add item", "charger", "Charger", "I-1", "1"},
	}
	for _, form := range forms {
		for end := 0; end < len(form); end++ {
			park, _ := runScript(t, strings.Join(form[:end], "\n")+"\n")
			if park.Summary() != (fleet.Summary{}) {
				t.Fatalf("partial form added an object: %v", form[:end])
			}
		}
	}
}

type failedReader struct{}

var errRead = errors.New("input unavailable")

func (failedReader) Read([]byte) (int, error) { return 0, errRead }

func TestReadFailureIsReturned(t *testing.T) {
	park := fleet.NewFleet(&fleet.ServiceCenter{})
	if err := cli.New(park, failedReader{}, io.Discard).Run(); !errors.Is(err, errRead) {
		t.Fatalf("read failure was hidden: %v", err)
	}
	for _, prefix := range []string{"", "add item\nhelmet\n"} {
		input := strings.NewReader(prefix + strings.Repeat("x", 70*1024))
		if err := cli.New(park, input, io.Discard).Run(); err == nil {
			t.Fatal("oversized input must return a read error")
		}
	}
}

func TestDemoCoversAdmissionAndReports(t *testing.T) {
	park := fleet.NewFleet(&fleet.ServiceCenter{})
	var output bytes.Buffer
	if err := cli.New(park, strings.NewReader(""), &output).Demo(); err != nil {
		t.Fatal(err)
	}
	summary := park.Summary()
	if summary.TransportCount != 4 || summary.ThingCount != 4 || math.Abs(summary.DailyEnergy-7) > 1e-9 {
		t.Fatalf("demo rejection or duplicate changed state: %+v", summary)
	}
	if len(park.ForBeginners()) != 3 {
		t.Fatal("demo must show both sides of the beginner threshold")
	}
	for _, expected := range []string{"Отказ техосмотра", "Повторный номер", "Ошибка создания", "Электросёрф Волна [T-004]", "Суточный расход энергии: 7.00"} {
		if !strings.Contains(output.String(), expected) {
			t.Errorf("demo did not show %q", expected)
		}
	}
}
