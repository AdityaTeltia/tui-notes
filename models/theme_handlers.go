package models

func (m *MainModel) cycleTheme() {
	themeNames := []string{"default", "dark", "light", "monokai", "nord"}
	
	currentIdx := 0
	for i, name := range themeNames {
		if name == m.currentTheme {
			currentIdx = i
			break
		}
	}
	
	nextIdx := (currentIdx + 1) % len(themeNames)
	m.ApplyTheme(themeNames[nextIdx])
}

