// Пакет fleet реализует приёмку и отчёты независимо от консольного интерфейса.
package fleet

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"

	"hw1/internal/model"
)

// Inspector объявлен в пакете потребителя, Fleet, и подменяется в тестах.
type Inspector interface {
	Inspect(model.Transport) bool
}

type ServiceCenter struct{}

func (ServiceCenter) Inspect(transport model.Transport) bool {
	return !transport.HasCriticalDamage()
}

var (
	ErrDuplicateNumber    = errors.New("инвентарный номер уже занят")
	ErrInspectionRejected = errors.New("техосмотр отклонил транспорт")
	ErrInvalidItem        = errors.New("некорректные данные имущества")
)

type Fleet struct {
	inspector      Inspector
	items          map[string]model.InventoryItem
	dailyEnergy    float64
	transportCount int
}

func NewFleet(inspector Inspector) *Fleet {
	return &Fleet{inspector: inspector, items: make(map[string]model.InventoryItem)}
}

// Add является единственным путём приёмки: транспорт проверяется даже при
// передаче как InventoryItem. Неудачная операция не меняет состояние парка.
func (park *Fleet) Add(item model.InventoryItem) error {
	if item == nil || (reflect.ValueOf(item).Kind() == reflect.Pointer && reflect.ValueOf(item).IsNil()) {
		return fmt.Errorf("%w: объект отсутствует", ErrInvalidItem)
	}
	number := item.InventoryNumber()
	if strings.TrimSpace(number) == "" || strings.TrimSpace(item.Name()) == "" {
		return fmt.Errorf("%w: нужны название и номер", ErrInvalidItem)
	}
	if _, exists := park.items[number]; exists {
		return fmt.Errorf("%w: %s", ErrDuplicateNumber, number)
	}

	energy := 0.0
	if consumer, ok := item.(model.EnergyConsumer); ok {
		energy = consumer.DailyEnergy()
		if energy < 0 || math.IsNaN(energy) || math.IsInf(energy, 0) || math.IsInf(park.dailyEnergy+energy, 0) {
			return fmt.Errorf("%w: недопустимый расход энергии", ErrInvalidItem)
		}
	}

	transport, isTransport := item.(model.Transport)
	if isTransport {
		if transport.Ease() < 1 || transport.Ease() > 10 {
			return fmt.Errorf("%w: простота должна быть от 1 до 10", ErrInvalidItem)
		}
		if !park.inspector.Inspect(transport) {
			return fmt.Errorf("%w: %s", ErrInspectionRejected, number)
		}
	}

	park.items[number] = item
	park.dailyEnergy += energy
	if isTransport {
		park.transportCount++
	}
	return nil
}

type Summary struct {
	TransportCount int
	ThingCount     int
	TotalItems     int
	DailyEnergy    float64
}

// Summary использует показатели, обновляемые только после успешной приёмки.
func (park *Fleet) Summary() Summary {
	return Summary{
		TransportCount: park.transportCount,
		ThingCount:     len(park.items) - park.transportCount,
		TotalItems:     len(park.items),
		DailyEnergy:    park.dailyEnergy,
	}
}

// Inventory возвращает новый срез, отсортированный по инвентарному номеру.
func (park *Fleet) Inventory() []model.InventoryItem {
	items := make([]model.InventoryItem, 0, len(park.items))
	for _, item := range park.items {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].InventoryNumber() < items[j].InventoryNumber()
	})
	return items
}

func (park *Fleet) ForBeginners() []model.Transport {
	transports := make([]model.Transport, 0)
	for _, item := range park.items {
		if transport, ok := item.(model.Transport); ok && transport.Ease() >= 6 {
			transports = append(transports, transport)
		}
	}
	sort.Slice(transports, func(i, j int) bool {
		if transports[i].Ease() == transports[j].Ease() {
			return transports[i].InventoryNumber() < transports[j].InventoryNumber()
		}
		return transports[i].Ease() > transports[j].Ease()
	})
	return transports
}
