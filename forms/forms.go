package forms

import (
	"github.com/charmbracelet/huh"
	"github.com/jon4hz/esi/config"
)

func TokenInputForm(token *string) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("TSS API Token").
				Prompt("? ").
				Password(true).
				Value(token),
		),
	)
}

func PasswordInputForm(password *string) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Local encryption password").
				Prompt("? ").
				Password(true).
				Value(password),
		),
	)
}

func GroupSelectForm(groups []*config.Group, out **config.Group) *huh.Form {
	options := make([]huh.Option[*config.Group], 0, len(groups))
	for _, g := range groups {
		options = append(options, huh.NewOption(g.Name, g).Selected(g.Selected))
	}
	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[*config.Group]().
				Options(options...).
				Title("Select Group").
				Value(out),
		),
	)
}

func InjectorSelectForm(injectors []*config.Injector, out **config.Injector) *huh.Form {
	options := make([]huh.Option[*config.Injector], 0, len(injectors))
	for _, inj := range injectors {
		options = append(options, huh.NewOption(inj.Name, inj).Selected(inj.Selected))
	}
	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[*config.Injector]().
				Options(options...).
				Title("Select Injectors").
				Value(out),
		),
	)
}
