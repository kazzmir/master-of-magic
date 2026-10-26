package engine

import (
	"fmt"
	"image/color"
	"slices"
	"strconv"

	uiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kazzmir/master-of-magic/game/magic/data"
	"github.com/kazzmir/master-of-magic/game/magic/setup"
	"github.com/kazzmir/master-of-magic/game/magic/spellbook"
	"github.com/kazzmir/master-of-magic/game/magic/util"
	"github.com/kazzmir/master-of-magic/game/network-combat/protocol"
	"github.com/kazzmir/master-of-magic/lib/lbx"
)

type gameSetupUI struct {
	match                                                         *matchmakingUI
	container                                                     *widget.Container
	budget, mana                                                  *widget.TextInput
	books                                                         [5]*widget.TextInput
	bookIcons                                                     [5]*widget.Graphic
	bookIconsReady                                                bool
	role, apply, save, addUnit, removeUnit, addSpell, removeSpell *widget.Button
	unitChoice, spellChoice                                       *widget.ListComboButton
	portraitButton                                                *portraitPreview
	portraitWindow                                                *widget.Window
	portraitGrid                                                  *widget.Container
	portraitChoices                                               []*portraitPreview
	army, spells                                                  *widget.List
	spent                                                         *widget.Text
	images                                                        map[int]*ebiten.Image
	imageCache                                                    *util.ImageCache
	lookedForAssets                                               bool
	draft                                                         protocol.Loadout
	gameID, attackerID                                            uint64
	dirty, settingsDirty, invalidText                             bool
}

func spellSetupMana(spell protocol.SpellChoice, books [5]uint64) int {
	wizard := &setup.WizardCustom{}
	for i, count := range books {
		wizard.Books = append(wizard.Books, data.WizardBook{Magic: data.MagicType(i + 1), Count: int(min(count, 11))})
	}
	magic := data.ArcaneMagic
	if index := protocol.MagicBookIndex(spell.Magic); index >= 0 {
		magic = data.MagicType(index + 1)
	}
	return spellbook.ComputeSpellCost(wizard, spellbook.Spell{CastCost: int(spell.Mana), Magic: magic}, false, false)
}

func spellSetupLabel(name string, books [5]uint64) string {
	spell, ok := protocol.FindSpell(name)
	if !ok {
		return name
	}
	return fmt.Sprintf("%s %d mana (%d)", spell.Name, spellSetupMana(spell, books), spell.Cost)
}

func (editor *gameSetupUI) knownSpellColor(name string) color.Color {
	spell, ok := protocol.FindSpell(name)
	if ok && uint64(spellSetupMana(spell, editor.draft.Books)) > editor.draft.Mana {
		return color.NRGBA{R: 255, G: 220, B: 80, A: 255}
	}
	return nil
}

func setupRow(weights []float64) *widget.Container {
	return widget.NewContainer(widget.ContainerOpts.Layout(widget.NewGridLayout(
		widget.GridLayoutOpts.Columns(len(weights)), widget.GridLayoutOpts.Weights(weights, []float64{1}), widget.GridLayoutOpts.Spacing(6, 4),
	)))
}

func newGameSetupUI(match *matchmakingUI, theme *widget.Theme) *gameSetupUI {
	editor := &gameSetupUI{match: match, images: make(map[int]*ebiten.Image)}
	label := func(value string) *widget.Text { return widget.NewText(widget.TextOpts.Text(value, nil, color.White)) }
	button := func(value string, action func()) *widget.Button {
		return widget.NewButton(widget.ButtonOpts.Text(value, nil, nil), widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) { action() }))
	}
	number := func(opts ...widget.WidgetOpt) *widget.TextInput {
		return widget.NewTextInput(widget.TextInputOpts.WidgetOpts(opts...), widget.TextInputOpts.Validation(func(value string) (bool, *string) {
			if value == "" {
				return true, nil
			}
			_, err := strconv.ParseUint(value, 10, 64)
			return err == nil, nil
		}))
	}
	combo := func(entries []any, text func(any) string, changed func(any)) *widget.ListComboButton {
		return widget.NewListComboButton(widget.ListComboButtonOpts.Entries(entries), widget.ListComboButtonOpts.MaxContentHeight(180),
			widget.ListComboButtonOpts.EntryLabelFunc(text, text), widget.ListComboButtonOpts.EntrySelectedHandler(func(args *widget.ListComboButtonEntrySelectedEventArgs) { changed(args.Entry) }))
	}
	editor.container = lobbyColumn(0, 1, 0)
	editor.container.GetWidget().SetTheme(theme)
	editor.budget = number()
	editor.budget.ChangedEvent.AddHandler(func(any) { editor.updateSettingsDirty() })
	editor.role = button("Attacker", func() {
		players := match.active.Players
		for i, p := range players {
			if p.ClientID == editor.attackerID {
				editor.attackerID = players[(i+1)%len(players)].ClientID
				break
			}
		}
		editor.updateSettingsDirty()
	})
	editor.apply = button("Apply settings", func() {
		budget, err := strconv.ParseUint(editor.budget.GetText(), 10, 64)
		if err != nil {
			match.game.status.Label = "Enter a uint64 budget"
			return
		}
		match.send(protocol.Message{Type: "game-settings", GameID: match.active.ID, Budget: budget, AttackerID: editor.attackerID})
	})
	settings := setupRow([]float64{0, 1, 1, 0})
	settings.AddChild(label("Budget per player"), editor.budget, editor.role, editor.apply)
	editor.container.AddChild(settings)
	columns := setupRow([]float64{0.34, 0.36, 0.3})
	// Each column's list gets the space left after its controls.
	armyPanel, spellPanel, wizardPanel := lobbyColumn(0, 0, 0, 1), lobbyColumn(0, 0, 0, 1), lobbyColumn(1, 0, 0)
	unitEntries := []any{}
	for _, unit := range protocol.UnitsCatalog() {
		unitEntries = append(unitEntries, unit.ID)
	}
	editor.unitChoice = combo(unitEntries, func(entry any) string {
		unit, _ := protocol.FindUnit(entry.(protocol.UnitID))
		return fmt.Sprintf("%s (%d)", unit.Name, unit.Cost)
	}, func(any) {})
	editor.army = widget.NewList(widget.ListOpts.EntryLabelFunc(func(entry any) string {
		i := entry.(int)
		unit, _ := protocol.FindUnit(editor.draft.Army[i])
		return fmt.Sprintf("%s (%d)", unit.Name, unit.Cost)
	}))
	editor.addUnit = button("Add unit", func() {
		if id, ok := editor.unitChoice.SelectedEntry().(protocol.UnitID); ok && len(editor.draft.Army) < protocol.MaxArmy {
			editor.draft.Army = append(editor.draft.Army, id)
			editor.dirty = true
			editor.refreshLists()
			editor.refresh()
		}
	})
	editor.removeUnit = button("Remove", func() {
		if i, ok := editor.army.SelectedEntry().(int); ok && i < len(editor.draft.Army) {
			editor.draft.Army = slices.Delete(editor.draft.Army, i, i+1)
			editor.dirty = true
			editor.refreshLists()
			editor.refresh()
		}
	})
	armyButtons := setupRow([]float64{1, 1})
	armyButtons.AddChild(editor.addUnit, editor.removeUnit)
	armyPanel.AddChild(label(fmt.Sprintf("Army (1–%d units)", protocol.MaxArmy)), editor.unitChoice, armyButtons, editor.army)
	spellEntries := []any{}
	for _, spell := range protocol.SpellsCatalog() {
		spellEntries = append(spellEntries, spell.Name)
	}
	editor.spellChoice = combo(spellEntries, func(entry any) string { return spellSetupLabel(entry.(string), editor.draft.Books) }, func(any) {})
	editor.spells = widget.NewList(
		widget.ListOpts.EntryLabelFunc(func(entry any) string { return spellSetupLabel(entry.(string), editor.draft.Books) }),
		widget.ListOpts.EntryTextColorFunc(func(entry any) color.Color { return editor.knownSpellColor(entry.(string)) }),
	)
	editor.addSpell = button("Learn spell", func() {
		if name, ok := editor.spellChoice.SelectedEntry().(string); ok && !slices.Contains(editor.draft.Spells, name) {
			editor.draft.Spells = append(editor.draft.Spells, name)
			editor.dirty = true
			editor.refreshLists()
			editor.refresh()
		}
	})
	editor.removeSpell = button("Forget", func() {
		if name, ok := editor.spells.SelectedEntry().(string); ok {
			editor.draft.Spells = slices.DeleteFunc(editor.draft.Spells, func(s string) bool { return s == name })
			editor.dirty = true
			editor.refreshLists()
			editor.refresh()
		}
	})
	spellButtons := setupRow([]float64{1, 1})
	spellButtons.AddChild(editor.addSpell, editor.removeSpell)
	spellPanel.AddChild(label("Known combat spells"), editor.spellChoice, spellButtons, editor.spells)
	editor.portraitButton = newPortraitPreview(ebiten.NewImage(64, 80), editor.openPortraitPicker)
	editor.mana = number()
	editor.mana.ChangedEvent.AddHandler(func(any) { editor.readNumbers() })
	manaColumn := lobbyColumn(0, 0)
	manaColumn.AddChild(label(fmt.Sprintf("Mana (%d each)", protocol.ManaCost)), editor.mana)
	portraitRow := setupRow([]float64{1, 1})
	portraitRow.AddChild(editor.portraitButton, manaColumn)
	books := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewGridLayout(
		widget.GridLayoutOpts.Columns(5), widget.GridLayoutOpts.Weights([]float64{1, 1, 1, 1, 1}, nil), widget.GridLayoutOpts.Spacing(2, 0),
	)))
	for i, name := range protocol.MagicNames {
		editor.books[i] = number(widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}))
		editor.books[i].ChangedEvent.AddHandler(func(any) { editor.readNumbers() })
		tipContent := lobbyColumn(0)
		tipContent.GetWidget().SetTheme(theme)
		tipContent.SetBackgroundImage(uiimage.NewNineSliceColor(color.NRGBA{R: 42, G: 50, B: 64, A: 255}))
		tipContent.AddChild(label(name))
		editor.bookIcons[i] = widget.NewGraphic(widget.GraphicOpts.Image(ebiten.NewImage(24, 24)),
			widget.GraphicOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{Position: widget.RowLayoutPositionCenter}),
				widget.WidgetOpts.ToolTip(widget.NewToolTip(widget.ToolTipOpts.Content(tipContent)))))
		column := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical), widget.RowLayoutOpts.Spacing(2))))
		column.AddChild(editor.bookIcons[i], editor.books[i])
		books.AddChild(column)
	}
	wizardPanel.AddChild(portraitRow, label(fmt.Sprintf("Books (%d each; max 11)", protocol.BookCost)), books)
	columns.AddChild(armyPanel, spellPanel, wizardPanel)
	editor.container.AddChild(columns)
	editor.spent = label("")
	editor.save = button("Save loadout", func() {
		editor.readNumbers()
		if editor.invalidText {
			return
		}
		loadout := editor.draft.Clone()
		match.send(protocol.Message{Type: "game-loadout", GameID: match.active.ID, Revision: match.active.Revision, Loadout: &loadout})
	})
	footer := setupRow([]float64{1, 0})
	footer.AddChild(editor.spent, editor.save)
	editor.container.AddChild(footer)
	return editor
}

func (editor *gameSetupUI) localPlayer() (protocol.GamePlayer, bool) {
	for _, p := range editor.match.active.Players {
		if p.ClientID == editor.match.game.clientID {
			return p, true
		}
	}
	return protocol.GamePlayer{}, false
}
func (editor *gameSetupUI) updateSettingsDirty() {
	value, err := strconv.ParseUint(editor.budget.GetText(), 10, 64)
	editor.settingsDirty = err != nil || value != editor.match.active.Budget || editor.attackerID != editor.match.active.AttackerID
	name := clientLabel(editor.attackerID)
	for _, p := range editor.match.active.Players {
		if p.ClientID == editor.attackerID {
			name = p.Name
		}
	}
	editor.role.SetText("Attacker: " + name)
	editor.match.setEnabled()
}
func (editor *gameSetupUI) readNumbers() {
	loadout := editor.draft.Clone()
	mana, err := strconv.ParseUint(editor.mana.GetText(), 10, 64)
	if err != nil {
		editor.invalidText = true
		editor.refresh()
		return
	}
	loadout.Mana = mana
	for i, input := range editor.books {
		count, err := strconv.ParseUint(input.GetText(), 10, 64)
		if err != nil {
			editor.invalidText = true
			editor.refresh()
			return
		}
		loadout.Books[i] = count
	}
	editor.invalidText = false
	booksChanged := loadout.Books != editor.draft.Books
	manaChanged := loadout.Mana != editor.draft.Mana
	if !loadout.Equal(editor.draft) {
		editor.draft = loadout
		editor.dirty = true
	}
	if booksChanged {
		editor.refreshSpellLabels()
	} else if manaChanged {
		editor.refreshKnownSpellRows()
	}
	editor.refresh()
}
func (editor *gameSetupUI) refreshSpellLabels() {
	editor.spellChoice.RefreshEntryLabels()
	editor.refreshKnownSpellRows()
}
func (editor *gameSetupUI) refreshKnownSpellRows() {
	for _, entry := range editor.spells.Entries() {
		editor.spells.UpdateEntry(entry)
	}
}
func (editor *gameSetupUI) refreshLists() {
	entries := make([]any, len(editor.draft.Army))
	for i := range entries {
		entries[i] = i
	}
	editor.army.SetEntries(entries)
	editor.spells.SetEntries(stringEntries(editor.draft.Spells))
	editor.spellChoice.RefreshEntryLabels()
}
func (editor *gameSetupUI) refresh() {
	spent, err := protocol.ValidateLoadout(editor.draft, editor.match.active.Budget, false)
	editor.spent.Label = fmt.Sprintf("Spent %d / %d", spent, editor.match.active.Budget)
	if err != nil {
		editor.spent.Label = err.Error()
	}
	if editor.invalidText {
		editor.spent.Label = "Enter valid mana and book counts"
	}
	if editor.dirty {
		editor.spent.Label += " (unsaved)"
	}
	editor.match.setEnabled()
}
func (editor *gameSetupUI) sync() {
	state := editor.match.active
	own, ok := editor.localPlayer()
	if !ok {
		return
	}
	newGame := editor.gameID != state.ID
	if !slices.ContainsFunc(state.Players, func(p protocol.GamePlayer) bool { return p.ClientID == editor.attackerID }) {
		editor.attackerID = state.AttackerID
	}
	if newGame || !editor.settingsDirty {
		editor.attackerID = state.AttackerID
		editor.budget.SetText(strconv.FormatUint(state.Budget, 10))
		editor.settingsDirty = false
	}
	if newGame || !editor.dirty || own.Loadout.Equal(editor.draft) {
		editor.draft = own.Loadout.Clone()
		editor.dirty = false
		editor.invalidText = false
		editor.mana.SetText(strconv.FormatUint(editor.draft.Mana, 10))
		for i, input := range editor.books {
			input.SetText(strconv.FormatUint(editor.draft.Books[i], 10))
		}
		editor.refreshLists()
		editor.showPortrait()
	}
	editor.gameID = state.ID
	// An acknowledged host edit must no longer block Ready.
	if value, err := strconv.ParseUint(editor.budget.GetText(), 10, 64); err == nil && value == state.Budget && editor.attackerID == state.AttackerID {
		editor.settingsDirty = false
	}
	editor.updateSettingsDirty()
	editor.refresh()
}
func (editor *gameSetupUI) showPortrait() {
	if !editor.lookedForAssets {
		editor.lookedForAssets = true
		if cache := lbx.AutoCache(); cache != nil {
			imageCache := util.MakeImageCache(cache)
			editor.imageCache = &imageCache
		}
	}
	if editor.imageCache != nil && !editor.bookIconsReady {
		editor.loadBookIcons()
	}
	img := editor.portraitImageFor(editor.draft.Portrait)
	editor.portraitButton.portrait = img
}
func (editor *gameSetupUI) canReady() bool {
	if own, ok := editor.localPlayer(); ok && own.Ready {
		return true
	}
	_, err := protocol.ValidateLoadout(editor.draft, editor.match.active.Budget, true)
	return !editor.dirty && !editor.settingsDirty && !editor.invalidText && err == nil
}
func (editor *gameSetupUI) setEnabled() {
	match := editor.match
	own, _ := editor.localPlayer()
	editable := match.game.connected && match.active.ID != 0 && !match.active.Started && !match.pending && !match.inviting && !own.Ready
	if !editable && editor.portraitWindow != nil {
		editor.portraitWindow.Close()
	}
	host := editable && match.game.clientID == match.active.HostID
	editor.budget.GetWidget().Disabled = !host
	editor.role.GetWidget().Disabled = !host || len(match.active.Players) < 2
	editor.apply.GetWidget().Disabled = !host || !editor.settingsDirty
	for _, w := range []widget.HasWidget{editor.mana, editor.portraitButton, editor.unitChoice, editor.spellChoice, editor.army, editor.spells, editor.addUnit, editor.removeUnit, editor.addSpell, editor.removeSpell} {
		w.GetWidget().Disabled = !editable
	}
	for _, input := range editor.books {
		input.GetWidget().Disabled = !editable
	}
	_, err := protocol.ValidateLoadout(editor.draft, match.active.Budget, false)
	editor.save.GetWidget().Disabled = !editable || !editor.dirty || editor.invalidText || err != nil
}
