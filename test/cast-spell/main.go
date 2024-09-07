package main

import (
    "log"

    uilib "github.com/kazzmir/master-of-magic/game/magic/ui"
    "github.com/kazzmir/master-of-magic/game/magic/spellbook"
    "github.com/kazzmir/master-of-magic/lib/lbx"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

const ScreenWidth = 320
const ScreenHeight = 200

type Engine struct {
    Counter uint64
    UI *uilib.UI
    Cache *lbx.LbxCache
}

func NewEngine() (*Engine, error) {
    cache := lbx.AutoCache()
    engine := &Engine{
        Counter: 0,
        Cache: cache,
    }

    engine.UI = engine.MakeUI()
    return engine, nil
}

func (engine *Engine) MakeUI() *uilib.UI {
    allSpells, err := spellbook.ReadSpellsFromCache(engine.Cache)
    if err != nil {
        log.Printf("Unable to read spells: %v", err)
        allSpells = spellbook.Spells{}
    }

    spells := spellbook.Spells{}
    spells.AddSpell(allSpells.FindByName("War Bears"))
    spells.AddSpell(allSpells.FindByName("Guardian Spirit"))
    spells.AddSpell(allSpells.FindByName("Sprites"))
    spells.AddSpell(allSpells.FindByName("Magic Spirit"))
    spells.AddSpell(allSpells.FindByName("Great Wyrm"))
    spells.AddSpell(allSpells.FindByName("Gorgons"))
    spells.AddSpell(allSpells.FindByName("Arch Angel"))
    spells.AddSpell(allSpells.FindByName("Bless"))
    spells.AddSpell(allSpells.FindByName("Iron Skin"))

    spells = allSpells

    ui := &uilib.UI{
        Draw: func(ui *uilib.UI, screen *ebiten.Image) {
            ui.IterateElementsByLayer(func(element *uilib.UIElement) {
                if element.Draw != nil {
                    element.Draw(element, screen)
                }
            })
        },
    }
    ui.SetElementsFromArray(nil)

    more := spellbook.MakeSpellBookCastUI(ui, engine.Cache, spells, 60)
    ui.AddElements(more)

    return ui
}

func (engine *Engine) Update() error {
    engine.Counter += 1
    keys := make([]ebiten.Key, 0)
    keys = inpututil.AppendJustPressedKeys(keys)

    for _, key := range keys {
        if key == ebiten.KeyEscape || key == ebiten.KeyCapsLock {
            return ebiten.Termination
        }
    }

    engine.UI.StandardUpdate()

    return nil
}

func (engine *Engine) Draw(screen *ebiten.Image){
    engine.UI.Draw(engine.UI, screen)
}

func (engine *Engine) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    return ScreenWidth, ScreenHeight
}

func main(){
    log.SetFlags(log.Ldate | log.Lshortfile | log.Lmicroseconds)

    ebiten.SetWindowSize(ScreenWidth * 5, ScreenHeight * 5)
    ebiten.SetWindowTitle("page turn")
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

    engine, err := NewEngine()

    err = ebiten.RunGame(engine)
    if err != nil {
        log.Printf("Error: %v", err)
    }
}
