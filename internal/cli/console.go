// Пакет cli отвечает за текстовый ввод и вывод; правила учёта находятся в Fleet.
package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"hw1/internal/fleet"
	"hw1/internal/model"
)

type Console struct {
	park    *fleet.Fleet
	scanner *bufio.Scanner
	output  io.Writer
}

func New(park *fleet.Fleet, input io.Reader, output io.Writer) *Console {
	return &Console{park: park, scanner: bufio.NewScanner(input), output: output}
}

func (console *Console) Run() error {
	fmt.Fprintln(console.output, "ВышКат — Доедем до дедлайна")
	console.help()
	for {
		line, err := console.readLine("> ")
		if errors.Is(err, io.EOF) {
			fmt.Fprintln(console.output, "\nВвод завершён.")
			return nil
		}
		if err != nil {
			return err
		}
		command := strings.ToLower(strings.Join(strings.Fields(line), " "))
		switch command {
		case "":
			continue
		case "help":
			console.help()
		case "add transport", "add item":
			err = console.add(command == "add transport")
			if errors.Is(err, io.EOF) {
				fmt.Fprintln(console.output, "\nВвод завершён. Объект не добавлен.")
				return nil
			}
			if console.scanner.Err() != nil {
				return err
			}
			if err != nil {
				fmt.Fprintf(console.output, "Ошибка: %v\n", err)
			}
		case "report":
			console.report()
		case "beginners":
			console.beginners()
		case "inventory":
			console.inventory()
		case "exit":
			fmt.Fprintln(console.output, "До встречи в ВышКат!")
			return nil
		default:
			fmt.Fprintln(console.output, "Неизвестная команда. Введите help.")
		}
	}
}

func (console *Console) help() {
	fmt.Fprintln(console.output, `Команды:
  add transport — добавить транспорт после техосмотра
  add item      — добавить вещь
  report        — количество техники, вещей и энергия
  beginners     — транспорт с простотой >= 6
  inventory     — всё имущество с инвентарными номерами
  help          — справка
  exit          — выход
Данные хранятся в памяти до выхода. Демо: go run . -demo`)
}

func (console *Console) readLine(prompt string) (string, error) {
	fmt.Fprint(console.output, prompt)
	if !console.scanner.Scan() {
		if err := console.scanner.Err(); err != nil {
			return "", fmt.Errorf("ошибка чтения: %w", err)
		}
		return "", io.EOF
	}
	return strings.TrimSpace(console.scanner.Text()), nil
}

func (console *Console) readInt(prompt string) (int, error) {
	line, err := console.readLine(prompt)
	if err != nil {
		return 0, err
	}
	value, err := strconv.Atoi(line)
	if err != nil {
		return 0, fmt.Errorf("ожидалось целое число")
	}
	return value, nil
}

func (console *Console) readEnergy() (float64, error) {
	line, err := console.readLine("Расход энергии, кВт·ч в сутки: ")
	if err != nil {
		return 0, err
	}
	value, err := strconv.ParseFloat(strings.ReplaceAll(line, ",", "."), 64)
	if err != nil {
		return 0, fmt.Errorf("ожидалось число, например 0.5")
	}
	return value, nil
}

func (console *Console) readDamage() (bool, error) {
	line, err := console.readLine("Есть критические повреждения? (да/нет): ")
	if err != nil {
		return false, err
	}
	switch strings.ToLower(line) {
	case "да", "yes", "true", "1":
		return true, nil
	case "нет", "no", "false", "0":
		return false, nil
	default:
		return false, fmt.Errorf("укажите да или нет")
	}
}

func (console *Console) add(transport bool) error {
	var kind string
	var err error
	if transport {
		kind, err = console.readLine("Вид: scooter (самокат), bicycle (велосипед), ebike (электровелосипед), surf (электросёрф): ")
	} else {
		kind, err = console.readLine("Вид: helmet (шлем), charger (зарядный шкаф), vest (жилет), dryer (сушилка гидрокостюмов): ")
	}
	if err != nil {
		return err
	}
	kind = strings.ToLower(kind)
	if transport && kind != "scooter" && kind != "bicycle" && kind != "ebike" && kind != "surf" {
		return fmt.Errorf("неизвестный вид транспорта")
	}
	if !transport && kind != "helmet" && kind != "charger" && kind != "vest" && kind != "dryer" {
		return fmt.Errorf("неизвестный вид вещи")
	}
	name, err := console.readLine("Название: ")
	if err != nil {
		return err
	}
	number, err := console.readLine("Инвентарный номер: ")
	if err != nil {
		return err
	}
	var item model.InventoryItem
	if transport {
		item, err = console.createTransport(kind, name, number)
	} else {
		item, err = console.createThing(kind, name, number)
	}
	if err != nil {
		return err
	}
	if err := console.park.Add(item); err != nil {
		return err
	}
	fmt.Fprintf(console.output, "Добавлено: %s [%s]\n", item.Name(), item.InventoryNumber())
	return nil
}

func (console *Console) createTransport(kind, name, number string) (model.InventoryItem, error) {
	ease, err := console.readInt("Простота для новичка (1–10): ")
	if err != nil {
		return nil, err
	}
	damaged, err := console.readDamage()
	if err != nil {
		return nil, err
	}
	if kind == "bicycle" {
		return model.NewBicycle(name, number, ease, damaged)
	}
	energy, err := console.readEnergy()
	if err != nil {
		return nil, err
	}
	switch kind {
	case "scooter":
		return model.NewScooter(name, number, ease, damaged, energy)
	case "ebike":
		return model.NewElectricBicycle(name, number, ease, damaged, energy)
	default: // Допустимость вида уже проверена в add.
		return model.NewElectricSurfboard(name, number, ease, damaged, energy)
	}
}

func (console *Console) createThing(kind, name, number string) (model.InventoryItem, error) {
	switch kind {
	case "helmet":
		return model.NewHelmet(name, number)
	case "vest":
		return model.NewLifeJacket(name, number)
	default:
		energy, err := console.readEnergy()
		if err != nil {
			return nil, err
		}
		if kind == "charger" {
			return model.NewChargingCabinet(name, number, energy)
		}
		return model.NewWetsuitDryer(name, number, energy)
	}
}

func (console *Console) report() {
	summary := console.park.Summary()
	fmt.Fprintf(console.output, "Транспорт: %d; вещи: %d; всего имущества: %d\n", summary.TransportCount, summary.ThingCount, summary.TotalItems)
	fmt.Fprintf(console.output, "Суточный расход энергии: %.2f кВт·ч\n", summary.DailyEnergy)
}

func (console *Console) beginners() {
	fmt.Fprintln(console.output, "Транспорт для новичков (простота >= 6):")
	transports := console.park.ForBeginners()
	if len(transports) == 0 {
		fmt.Fprintln(console.output, "Подходящего транспорта нет.")
	}
	for _, transport := range transports {
		fmt.Fprintf(console.output, "  %s [%s], простота: %d\n", transport.Name(), transport.InventoryNumber(), transport.Ease())
	}
}

func (console *Console) inventory() {
	fmt.Fprintln(console.output, "Всё имущество (по инвентарному номеру):")
	items := console.park.Inventory()
	if len(items) == 0 {
		fmt.Fprintln(console.output, "Имущество отсутствует.")
	}
	for _, item := range items {
		category := "вещь"
		if _, ok := item.(model.Transport); ok {
			category = "транспорт"
		}
		fmt.Fprintf(console.output, "  %s [%s] — %s\n", item.Name(), item.InventoryNumber(), category)
	}
}
