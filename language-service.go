package main

import (
	"manjaro-control-panel/backend"
)

type LanguageService struct{}

func (g *LanguageService) Packages() []backend.Language_package {
	return backend.Get_language_packs()
}
