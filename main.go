package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"go.uber.org/dig"

	"hw1/internal/cli"
	"hw1/internal/fleet"
)

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка:", err)
		os.Exit(1)
	}
}

func run(args []string, input io.Reader, output io.Writer) error {
	flags := flag.NewFlagSet("vyshkat", flag.ContinueOnError)
	flags.SetOutput(output)
	demo := flags.Bool("demo", false, "показать все основные сценарии на новом парке")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("неожиданные аргументы; используйте -help")
	}
	container, err := buildContainer(input, output, func() fleet.Inspector {
		return &fleet.ServiceCenter{}
	})
	if err != nil {
		return err
	}
	return container.Invoke(func(console *cli.Console) error {
		if *demo {
			return console.Demo()
		}
		return console.Run()
	})
}

// Контейнер используется только при сборке приложения. Парк и консоль
// не импортируют Dig и не запрашивают зависимости самостоятельно.
func buildContainer(input io.Reader, output io.Writer, inspectorProvider func() fleet.Inspector) (*dig.Container, error) {
	container := dig.New()
	constructors := []any{
		inspectorProvider,
		func() io.Reader { return input },
		func() io.Writer { return output },
		fleet.NewFleet,
		cli.New,
	}
	for _, constructor := range constructors {
		if err := container.Provide(constructor); err != nil {
			return nil, fmt.Errorf("регистрация зависимости: %w", err)
		}
	}
	return container, nil
}
