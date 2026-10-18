// Пакет model описывает имущество через композицию и небольшие интерфейсы.
package model

import (
	"fmt"
	"math"
	"strings"
)

// InventoryItem задаёт общий контракт для всего имущества на балансе.
type InventoryItem interface {
	Name() string
	InventoryNumber() string
}

// EnergyConsumer сообщает заданный суточный расход энергии.
type EnergyConsumer interface {
	DailyEnergy() float64
}

// Transport добавляет характеристики для техосмотра и отбора новичкам.
type Transport interface {
	InventoryItem
	Ease() int
	HasCriticalDamage() bool
}

type inventoryData struct {
	name   string
	number string
}

func (data inventoryData) Name() string            { return data.name }
func (data inventoryData) InventoryNumber() string { return data.number }

func newInventoryData(name, number string) (inventoryData, error) {
	name, number = strings.TrimSpace(name), strings.TrimSpace(number)
	if name == "" {
		return inventoryData{}, fmt.Errorf("название не должно быть пустым")
	}
	if number == "" {
		return inventoryData{}, fmt.Errorf("инвентарный номер не должен быть пустым")
	}
	return inventoryData{name: name, number: number}, nil
}

type transportData struct {
	inventoryData
	ease           int
	criticalDamage bool
}

func (data transportData) Ease() int               { return data.ease }
func (data transportData) HasCriticalDamage() bool { return data.criticalDamage }

func newTransportData(name, number string, ease int, damaged bool) (transportData, error) {
	item, err := newInventoryData(name, number)
	if err != nil {
		return transportData{}, err
	}
	if ease < 1 || ease > 10 {
		return transportData{}, fmt.Errorf("простота должна быть от 1 до 10")
	}
	return transportData{inventoryData: item, ease: ease, criticalDamage: damaged}, nil
}

type energyData struct {
	dailyEnergy float64
}

func (data energyData) DailyEnergy() float64 { return data.dailyEnergy }

func newEnergyData(energy float64) (energyData, error) {
	if energy < 0 || math.IsNaN(energy) || math.IsInf(energy, 0) {
		return energyData{}, fmt.Errorf("расход энергии должен быть конечным числом не меньше нуля")
	}
	return energyData{dailyEnergy: energy}, nil
}

// Конструкторы возвращают значения без методов изменения характеристик.
// При передаче значений парк хранит копии, независимые от переменных вызывающего кода.
type Bicycle struct{ transportData }

func NewBicycle(name, number string, ease int, damaged bool) (Bicycle, error) {
	data, err := newTransportData(name, number, ease, damaged)
	return Bicycle{data}, err
}

type Scooter struct {
	transportData
	energyData
}

func NewScooter(name, number string, ease int, damaged bool, energy float64) (Scooter, error) {
	transport, consumption, err := newPoweredTransport(name, number, ease, damaged, energy)
	return Scooter{transport, consumption}, err
}

type ElectricBicycle struct {
	transportData
	energyData
}

func NewElectricBicycle(name, number string, ease int, damaged bool, energy float64) (ElectricBicycle, error) {
	transport, consumption, err := newPoweredTransport(name, number, ease, damaged, energy)
	return ElectricBicycle{transport, consumption}, err
}

type ElectricSurfboard struct {
	transportData
	energyData
}

func NewElectricSurfboard(name, number string, ease int, damaged bool, energy float64) (ElectricSurfboard, error) {
	transport, consumption, err := newPoweredTransport(name, number, ease, damaged, energy)
	return ElectricSurfboard{transport, consumption}, err
}

func newPoweredTransport(name, number string, ease int, damaged bool, energy float64) (transportData, energyData, error) {
	transport, err := newTransportData(name, number, ease, damaged)
	if err != nil {
		return transportData{}, energyData{}, err
	}
	consumption, err := newEnergyData(energy)
	return transport, consumption, err
}

type Helmet struct{ inventoryData }

func NewHelmet(name, number string) (Helmet, error) {
	data, err := newInventoryData(name, number)
	return Helmet{data}, err
}

type LifeJacket struct{ inventoryData }

func NewLifeJacket(name, number string) (LifeJacket, error) {
	data, err := newInventoryData(name, number)
	return LifeJacket{data}, err
}

type ChargingCabinet struct {
	inventoryData
	energyData
}

func NewChargingCabinet(name, number string, energy float64) (ChargingCabinet, error) {
	item, consumption, err := newPoweredItem(name, number, energy)
	return ChargingCabinet{item, consumption}, err
}

type WetsuitDryer struct {
	inventoryData
	energyData
}

func NewWetsuitDryer(name, number string, energy float64) (WetsuitDryer, error) {
	item, consumption, err := newPoweredItem(name, number, energy)
	return WetsuitDryer{item, consumption}, err
}

func newPoweredItem(name, number string, energy float64) (inventoryData, energyData, error) {
	item, err := newInventoryData(name, number)
	if err != nil {
		return inventoryData{}, energyData{}, err
	}
	consumption, err := newEnergyData(energy)
	return item, consumption, err
}
