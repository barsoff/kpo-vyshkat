package cli

import (
	"errors"
	"fmt"

	"hw1/internal/fleet"
	"hw1/internal/model"
)

// Demo запускается с новым парком, собранным тем же контейнером, что и обычный CLI.
func (console *Console) Demo() error {
	fmt.Fprintln(console.output, "ВышКат — демонстрация сценариев")
	fmt.Fprintln(console.output, "\n1. Пустой парк")
	console.report()
	console.beginners()
	console.inventory()

	fmt.Fprintln(console.output, "\n2. Приём исправного транспорта и вещей")
	if err := console.demoAdd(model.NewBicycle("Велосипед Кампус", "T-001", 6, false)); err != nil {
		return err
	}
	if err := console.demoAdd(model.NewScooter("Самокат Дедлайн", "T-002", 8, false, 0.5)); err != nil {
		return err
	}
	if err := console.demoAdd(model.NewElectricBicycle("Электровелосипед ФКН", "T-003", 6, false, 0.8)); err != nil {
		return err
	}
	if err := console.demoAdd(model.NewElectricSurfboard("Электросёрф Волна", "T-004", 5, false, 2.2)); err != nil {
		return err
	}
	if err := console.demoAdd(model.NewHelmet("Шлем Защита", "I-001")); err != nil {
		return err
	}
	if err := console.demoAdd(model.NewChargingCabinet("Зарядный шкаф", "I-002", 1.5)); err != nil {
		return err
	}
	if err := console.demoAdd(model.NewLifeJacket("Спасательный жилет", "I-003")); err != nil {
		return err
	}
	if err := console.demoAdd(model.NewWetsuitDryer("Сушилка гидрокостюмов", "I-004", 2)); err != nil {
		return err
	}

	fmt.Fprintln(console.output, "\n3. Отказ техосмотра")
	damaged, err := model.NewScooter("Повреждённый самокат", "T-005", 9, true, 0.9)
	if err != nil {
		return err
	}
	if err := console.park.Add(damaged); !errors.Is(err, fleet.ErrInspectionRejected) {
		return fmt.Errorf("ожидался отказ техосмотра, получено: %v", err)
	} else {
		fmt.Fprintf(console.output, "Ожидаемый отказ: %v\n", err)
	}

	fmt.Fprintln(console.output, "\n4. Повторный номер техники у вещи")
	duplicate, err := model.NewHelmet("Другой шлем", "T-002")
	if err != nil {
		return err
	}
	if err := console.park.Add(duplicate); !errors.Is(err, fleet.ErrDuplicateNumber) {
		return fmt.Errorf("ожидался отказ при повторном номере, получено: %v", err)
	} else {
		fmt.Fprintf(console.output, "Ожидаемый отказ: %v\n", err)
	}

	fmt.Fprintln(console.output, "\n5. Некорректные исходные данные")
	if _, err := model.NewBicycle("Велосипед", "T-006", 11, false); err != nil {
		fmt.Fprintf(console.output, "Ошибка создания: %v\n", err)
	} else {
		return fmt.Errorf("ожидалась ошибка при простоте 11")
	}

	fmt.Fprintln(console.output, "\n6. Итоговые отчёты; отказы не изменили парк")
	console.report()
	console.beginners()
	console.inventory()
	return nil
}

func (console *Console) demoAdd(item model.InventoryItem, err error) error {
	if err != nil {
		return err
	}
	if err := console.park.Add(item); err != nil {
		return err
	}
	fmt.Fprintf(console.output, "Добавлено: %s [%s]\n", item.Name(), item.InventoryNumber())
	return nil
}
