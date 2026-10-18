package fleet_test

import (
	"errors"
	"math"
	"reflect"
	"testing"

	"hw1/internal/fleet"
	"hw1/internal/model"
)

type stubInspector struct {
	accept bool
	calls  int
	lastID string
}

func (fake *stubInspector) Inspect(transport model.Transport) bool {
	fake.calls++
	fake.lastID = transport.InventoryNumber()
	return fake.accept
}

func bicycle(t *testing.T, id string, ease int, damaged bool) model.Bicycle {
	t.Helper()
	item, err := model.NewBicycle("Велосипед", id, ease, damaged)
	if err != nil {
		t.Fatal(err)
	}
	return item
}

func scooter(t *testing.T, id string, damaged bool, energy float64) model.Scooter {
	t.Helper()
	item, err := model.NewScooter("Самокат", id, 8, damaged, energy)
	if err != nil {
		t.Fatal(err)
	}
	return item
}

func TestEmptyFleet(t *testing.T) {
	park := fleet.NewFleet(&fleet.ServiceCenter{})
	if park.Summary() != (fleet.Summary{}) || len(park.Inventory()) != 0 || len(park.ForBeginners()) != 0 {
		t.Fatal("новый парк должен быть пустым")
	}
}

func TestServiceCenterDecision(t *testing.T) {
	center := fleet.ServiceCenter{}
	cases := []struct {
		name    string
		damaged bool
		accept  bool
	}{
		{"healthy_transport_is_accepted", false, true},
		{"damaged_transport_is_rejected", true, false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			// Низкая простота не означает техническую неисправность.
			if got := center.Inspect(bicycle(t, "T-1", 1, test.damaged)); got != test.accept {
				t.Errorf("решение о приёмке: получено %v, ожидалось %v", got, test.accept)
			}
		})
	}
}

func TestAdmissionUsesInjectedInspector(t *testing.T) {
	for _, accept := range []bool{true, false} {
		t.Run(map[bool]string{true: "stub_accepts_damaged", false: "stub_rejects_healthy"}[accept], func(t *testing.T) {
			inspector := &stubInspector{accept: accept}
			park := fleet.NewFleet(inspector)
			err := park.Add(scooter(t, "T-1", accept, 0.5))
			if inspector.calls != 1 || inspector.lastID != "T-1" {
				t.Fatal("парк должен вызвать переданный инспектор для добавляемого транспорта")
			}
			if accept {
				if err != nil || park.Summary() != (fleet.Summary{TransportCount: 1, TotalItems: 1, DailyEnergy: 0.5}) {
					t.Fatalf("одобренный транспорт не учтён: ошибка %v, сводка %+v", err, park.Summary())
				}
			} else if !errors.Is(err, fleet.ErrInspectionRejected) || park.Summary() != (fleet.Summary{}) {
				t.Fatalf("не получен ожидаемый отказ техосмотра или изменилось состояние парка: ошибка %v, сводка %+v", err, park.Summary())
			}
		})
	}
}

func TestThingsSkipInspectionAndContributeEnergy(t *testing.T) {
	inspector := &stubInspector{accept: false}
	park := fleet.NewFleet(inspector)
	helmet, err := model.NewHelmet("Шлем", "I-1")
	if err != nil {
		t.Fatal(err)
	}
	charger, err := model.NewChargingCabinet("Шкаф", "I-2", 1.5)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range []model.InventoryItem{helmet, charger} {
		if err := park.Add(item); err != nil {
			t.Fatal(err)
		}
	}
	if inspector.calls != 0 || park.Summary() != (fleet.Summary{ThingCount: 2, TotalItems: 2, DailyEnergy: 1.5}) {
		t.Fatal("вещи должны учитываться вместе с их расходом энергии без вызова техосмотра транспорта")
	}
}

func TestDuplicateNumberDoesNotChangeState(t *testing.T) {
	inspector := &stubInspector{accept: true}
	park := fleet.NewFleet(inspector)
	if err := park.Add(scooter(t, "SAME", false, 0.5)); err != nil {
		t.Fatal(err)
	}
	before := park.Summary()
	thing, err := model.NewChargingCabinet("Шкаф", "SAME", 10)
	if err != nil {
		t.Fatal(err)
	}
	for _, duplicate := range []model.InventoryItem{thing, scooter(t, "SAME", false, 3)} {
		if err := park.Add(duplicate); !errors.Is(err, fleet.ErrDuplicateNumber) {
			t.Fatalf("ожидалась ошибка повторного номера, получено: %v", err)
		}
		if park.Summary() != before || park.Inventory()[0].Name() != "Самокат" {
			t.Fatal("попытка добавления повторного номера заменила объект или изменила сводку парка")
		}
	}
}

func TestEnergyAccumulatesOnlyForAcceptedInventory(t *testing.T) {
	park := fleet.NewFleet(&fleet.ServiceCenter{})
	charger, err := model.NewChargingCabinet("Шкаф", "I-1", 1.25)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range []model.InventoryItem{
		scooter(t, "T-1", false, 0.5), bicycle(t, "T-2", 5, false), charger,
	} {
		if err := park.Add(item); err != nil {
			t.Fatal(err)
		}
	}
	want := fleet.Summary{TransportCount: 2, ThingCount: 1, TotalItems: 3, DailyEnergy: 1.75}
	if got := park.Summary(); got != want {
		t.Fatalf("сводка по транспорту и вещам: получено %+v, ожидалось %+v", got, want)
	}
	if err := park.Add(scooter(t, "REJECTED", true, 10)); !errors.Is(err, fleet.ErrInspectionRejected) {
		t.Fatalf("ожидался отказ техосмотра, получено: %v", err)
	}
	if err := park.Add(charger); !errors.Is(err, fleet.ErrDuplicateNumber) {
		t.Fatalf("ожидалась ошибка повторного номера, получено: %v", err)
	}
	if got := park.Summary(); got != want {
		t.Fatalf("неудачная приёмка изменила сводку парка: %+v", got)
	}
	if got := park.Inventory(); len(got) != 3 || got[0].InventoryNumber() != "I-1" || got[1].InventoryNumber() != "T-1" || got[2].InventoryNumber() != "T-2" {
		t.Fatal("неудачная приёмка изменила список имущества")
	}
}

func TestRejectedNumberCanBeAcceptedLater(t *testing.T) {
	park := fleet.NewFleet(&fleet.ServiceCenter{})
	if err := park.Add(scooter(t, "T-1", true, 0.5)); !errors.Is(err, fleet.ErrInspectionRejected) {
		t.Fatal(err)
	}
	if err := park.Add(scooter(t, "T-1", false, 0.5)); err != nil {
		t.Fatalf("номер отклонённого транспорта должен оставаться свободным: %v", err)
	}
	if park.Summary().DailyEnergy != 0.5 || park.Summary().TransportCount != 1 {
		t.Fatal("неверное состояние парка после приёмки исправного транспорта с ранее отклонённым номером")
	}
}

func TestBeginnerBoundaryAndOrder(t *testing.T) {
	park := fleet.NewFleet(&fleet.ServiceCenter{})
	for _, input := range []struct {
		id   string
		ease int
	}{{"T-5", 5}, {"T-6b", 6}, {"T-10", 10}, {"T-6a", 6}, {"T-1", 1}} {
		if err := park.Add(bicycle(t, input.id, input.ease, false)); err != nil {
			t.Fatal(err)
		}
	}
	got := make([]string, 0)
	for _, transport := range park.ForBeginners() {
		got = append(got, transport.InventoryNumber())
	}
	if want := []string{"T-10", "T-6a", "T-6b"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("список транспорта для новичков: получено %v, ожидалось %v", got, want)
	}
}

// Неизвестный парку тип проверяет расширение учёта через интерфейсы.
type customTransport struct {
	id     string
	ease   int
	energy float64
}

func (item customTransport) Name() string            { return "Новый транспорт" }
func (item customTransport) InventoryNumber() string { return item.id }
func (item customTransport) Ease() int               { return item.ease }
func (item customTransport) HasCriticalDamage() bool { return false }
func (item customTransport) DailyEnergy() float64    { return item.energy }

func TestNewTransportWorksWithoutFleetChanges(t *testing.T) {
	park := fleet.NewFleet(&fleet.ServiceCenter{})
	if err := park.Add(customTransport{id: "NEW", ease: 6, energy: 0.25}); err != nil {
		t.Fatal(err)
	}
	if park.Summary().DailyEnergy != 0.25 || len(park.ForBeginners()) != 1 || park.Inventory()[0].InventoryNumber() != "NEW" {
		t.Fatal("новая реализация транспорта не учтена в отчётах")
	}
}

func TestInvalidAdmissionDoesNotChangeState(t *testing.T) {
	var nilScooter *model.Scooter
	invalid := []model.InventoryItem{
		nil, nilScooter, model.Helmet{}, model.Bicycle{},
		customTransport{id: "N", ease: 0}, customTransport{id: "N", ease: 11},
		customTransport{id: "N", ease: 6, energy: -1},
		customTransport{id: "N", ease: 6, energy: math.NaN()},
		customTransport{id: "N", ease: 6, energy: math.Inf(1)},
	}
	for _, item := range invalid {
		inspector := &stubInspector{accept: true}
		park := fleet.NewFleet(inspector)
		if err := park.Add(item); !errors.Is(err, fleet.ErrInvalidItem) {
			t.Errorf("для некорректного объекта типа %T ожидалась ошибка ErrInvalidItem, получено: %v", item, err)
		}
		if park.Summary() != (fleet.Summary{}) || inspector.calls != 0 {
			t.Fatal("попытка добавления некорректного объекта изменила состояние парка или вызвала техосмотр")
		}
	}
}

func TestEnergyOverflowIsRejected(t *testing.T) {
	park := fleet.NewFleet(&fleet.ServiceCenter{})
	if err := park.Add(customTransport{id: "A", ease: 6, energy: math.MaxFloat64}); err != nil {
		t.Fatal(err)
	}
	before := park.Summary()
	if err := park.Add(customTransport{id: "B", ease: 6, energy: math.MaxFloat64}); !errors.Is(err, fleet.ErrInvalidItem) {
		t.Fatalf("переполнение суммы энергии должно отклоняться с ошибкой ErrInvalidItem, получено: %v", err)
	}
	if park.Summary() != before {
		t.Fatal("попытка добавления с переполнением энергии изменила сводку парка")
	}
}

func TestInventoryOrderAndIndependence(t *testing.T) {
	park := fleet.NewFleet(&fleet.ServiceCenter{})
	item := bicycle(t, "B", 6, false)
	if err := park.Add(item); err != nil {
		t.Fatal(err)
	}
	item = bicycle(t, "A", 10, false)
	if err := park.Add(item); err != nil {
		t.Fatal(err)
	}
	items := park.Inventory()
	if items[0].InventoryNumber() != "A" || items[1].InventoryNumber() != "B" {
		t.Fatal("список имущества не отсортирован по инвентарному номеру")
	}
	items[0] = nil
	if park.Inventory()[0] == nil || park.Summary().TotalItems != 2 {
		t.Fatal("изменение возвращённого среза повлияло на имущество парка")
	}
}
