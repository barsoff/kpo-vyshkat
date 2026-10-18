package model_test

import (
	"math"
	"testing"

	"hw1/internal/model"
)

func TestTransportConstruction(t *testing.T) {
	factories := []struct {
		name    string
		powered bool
		create  func(string, string, int, bool, float64) (model.Transport, error)
	}{
		{"bicycle", false, func(name, id string, ease int, damage bool, _ float64) (model.Transport, error) {
			return model.NewBicycle(name, id, ease, damage)
		}},
		{"scooter", true, func(name, id string, ease int, damage bool, energy float64) (model.Transport, error) {
			return model.NewScooter(name, id, ease, damage, energy)
		}},
		{"ebike", true, func(name, id string, ease int, damage bool, energy float64) (model.Transport, error) {
			return model.NewElectricBicycle(name, id, ease, damage, energy)
		}},
		{"surf", true, func(name, id string, ease int, damage bool, energy float64) (model.Transport, error) {
			return model.NewElectricSurfboard(name, id, ease, damage, energy)
		}},
	}
	for _, factory := range factories {
		t.Run(factory.name, func(t *testing.T) {
			for _, ease := range []int{1, 5, 6, 10} {
				item, err := factory.create("  Транспорт  ", " T-1 ", ease, true, 0.5)
				if err != nil {
					t.Fatal(err)
				}
				if item.Name() != "Транспорт" || item.InventoryNumber() != "T-1" || item.Ease() != ease || !item.HasCriticalDamage() {
					t.Fatalf("конструктор не сохранил характеристики или не удалил пробелы по краям: %+v", item)
				}
				consumer, powered := item.(model.EnergyConsumer)
				if powered != factory.powered || (powered && consumer.DailyEnergy() != 0.5) {
					t.Fatal("неверная реализация интерфейса EnergyConsumer или значение расхода энергии")
				}
			}
			invalid := []struct {
				name string
				id   string
				ease int
			}{
				{" ", "T-1", 6}, {"Name", "\t", 6},
				{"Name", "T-1", 0}, {"Name", "T-1", 11},
			}
			for _, input := range invalid {
				if _, err := factory.create(input.name, input.id, input.ease, false, 0); err == nil {
					t.Errorf("конструктор принял недопустимые характеристики: %+v", input)
				}
			}
			if factory.powered {
				for _, energy := range []float64{-0.01, math.NaN(), math.Inf(1), math.Inf(-1)} {
					if _, err := factory.create("Name", "T-1", 6, false, energy); err == nil {
						t.Errorf("конструктор принял недопустимый расход энергии: %v", energy)
					}
				}
				if _, err := factory.create("Name", "T-1", 6, false, 0); err != nil {
					t.Fatalf("нулевой расход энергии должен быть допустим: %v", err)
				}
			}
		})
	}
}

func TestThingConstruction(t *testing.T) {
	factories := []struct {
		name    string
		powered bool
		create  func(string, string, float64) (model.InventoryItem, error)
	}{
		{"helmet", false, func(name, id string, _ float64) (model.InventoryItem, error) {
			return model.NewHelmet(name, id)
		}},
		{"vest", false, func(name, id string, _ float64) (model.InventoryItem, error) {
			return model.NewLifeJacket(name, id)
		}},
		{"charger", true, func(name, id string, energy float64) (model.InventoryItem, error) {
			return model.NewChargingCabinet(name, id, energy)
		}},
		{"dryer", true, func(name, id string, energy float64) (model.InventoryItem, error) {
			return model.NewWetsuitDryer(name, id, energy)
		}},
	}
	for _, factory := range factories {
		t.Run(factory.name, func(t *testing.T) {
			item, err := factory.create("  Вещь ", " I-1 ", 1.25)
			if err != nil {
				t.Fatal(err)
			}
			if item.Name() != "Вещь" || item.InventoryNumber() != "I-1" {
				t.Fatal("пробелы по краям названия и инвентарного номера не удалены")
			}
			if _, isTransport := item.(model.Transport); isTransport {
				t.Fatal("вещь не должна реализовывать интерфейс Transport")
			}
			consumer, powered := item.(model.EnergyConsumer)
			if powered != factory.powered || (powered && consumer.DailyEnergy() != 1.25) {
				t.Fatal("неверная реализация интерфейса EnergyConsumer или значение расхода энергии")
			}
			for _, input := range [][2]string{{" ", "I-1"}, {"Name", " "}} {
				if _, err := factory.create(input[0], input[1], 0); err == nil {
					t.Fatal("конструктор принял пустое название или инвентарный номер")
				}
			}
			if factory.powered {
				for _, energy := range []float64{-1, math.NaN(), math.Inf(1), math.Inf(-1)} {
					if _, err := factory.create("Name", "I-1", energy); err == nil {
						t.Errorf("конструктор принял недопустимый расход энергии: %v", energy)
					}
				}
				if _, err := factory.create("Name", "I-1", 0); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
