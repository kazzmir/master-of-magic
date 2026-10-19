package ui

import (
	"image"
	"testing"

	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

func TestColumnFillsWidth(t *testing.T) {
	column := Column()
	child := widget.NewGraphic(widget.GraphicOpts.Image(ebiten.NewImage(20, 10)))
	column.AddChild(child)
	column.SetLocation(image.Rect(0, 0, 300, 100))
	column.RequestRelayout()
	column.Render(ebiten.NewImage(300, 100))
	if got := child.GetWidget().Rect.Dx(); got != 288 {
		t.Fatalf("child width = %d, want 288 after padding", got)
	}
}

func TestRecruitmentViewportFillsRemainingHeight(t *testing.T) {
	shop := Column(0, 0, 1)
	controls := Column(0, 1)
	list := Column(1)
	catalog := Column(0, 0, 1)
	row := ScrollRow(1)
	graphic := func(width, height int) *widget.Graphic {
		return widget.NewGraphic(widget.GraphicOpts.Image(ebiten.NewImage(width, height)))
	}
	viewport := widget.NewScrollContainer(
		widget.ScrollContainerOpts.Content(graphic(100, 2000)),
		widget.ScrollContainerOpts.StretchContentWidth(),
		widget.ScrollContainerOpts.Image(&widget.ScrollContainerImage{
			Idle: SolidImage(32, 32, 32), Mask: SolidImage(255, 255, 255),
		}),
	)
	scrollbar := graphic(10, 10)
	row.AddChild(viewport, scrollbar)
	catalog.AddChild(graphic(100, 20), graphic(100, 20), row)
	list.AddChild(catalog)
	controls.AddChild(graphic(100, 20), list)
	shop.AddChild(graphic(100, 20), graphic(100, 20), controls)

	for _, height := range []int{600, 1000, 450} {
		shop.SetLocation(image.Rect(0, 0, 800, height))
		shop.Render(ebiten.NewImage(800, height))
		// Four padded columns separate the viewport from the tab's bottom.
		wantBottom := height - 4*6
		if got := viewport.GetWidget().Rect.Max.Y; got != wantBottom {
			t.Fatalf("window height %d: viewport bottom = %d, want %d", height, got, wantBottom)
		}
		if scrollbar.GetWidget().Rect.Min.Y != viewport.GetWidget().Rect.Min.Y || scrollbar.GetWidget().Rect.Max.Y != wantBottom {
			t.Fatalf("scrollbar does not track viewport: %v", scrollbar.GetWidget().Rect)
		}
	}
}

func TestArmyAndUpgradePanelsFillHeight(t *testing.T) {
	graphic := func(width, height int) *widget.Graphic {
		return widget.NewGraphic(widget.GraphicOpts.Image(ebiten.NewImage(width, height)))
	}
	root := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewGridLayout(
		widget.GridLayoutOpts.Columns(2),
		widget.GridLayoutOpts.Weights([]float64{2, 3}, []float64{1}),
		widget.GridLayoutOpts.Padding(widget.NewInsetsSimple(8)),
	)))
	army := Column(0, 1)
	armyRow := ScrollRow(1)
	armyViewport, armyBar := graphic(100, 2000), graphic(10, 10)
	armyRow.AddChild(armyViewport, armyBar)
	army.AddChild(graphic(100, 20), armyRow)

	details := Column(0, 1)
	upgrades := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewGridLayout(
		widget.GridLayoutOpts.Columns(2),
		widget.GridLayoutOpts.Weights([]float64{1, 2}, []float64{1}),
	)))
	applied, available := Column(0, 1), Column(0, 1)
	appliedViewport := graphic(100, 2000)
	availableViewport, upgradeBar := graphic(100, 2000), graphic(10, 10)
	upgradeRow := ScrollRow(1)
	upgradeRow.AddChild(availableViewport, upgradeBar)
	applied.AddChild(graphic(100, 20), appliedViewport)
	available.AddChild(graphic(100, 20), upgradeRow)
	upgrades.AddChild(applied, available)
	details.AddChild(graphic(100, 220), upgrades)
	root.AddChild(army, details)

	for _, height := range []int{600, 1000, 500} {
		root.SetLocation(image.Rect(0, 0, 1200, height))
		root.Render(ebiten.NewImage(1200, height))
		for _, panel := range []struct {
			name   string
			widget widget.PreferredSizeLocateableWidget
			bottom int
		}{
			{"army", armyViewport, height - 14},
			{"army scrollbar", armyBar, height - 14},
			{"applied enchantments", appliedViewport, height - 20},
			{"available enchantments", availableViewport, height - 20},
			{"upgrade scrollbar", upgradeBar, height - 20},
		} {
			if got := panel.widget.GetWidget().Rect.Max.Y; got != panel.bottom {
				t.Errorf("height %d: %s bottom = %d, want %d", height, panel.name, got, panel.bottom)
			}
		}
	}
}

func TestScrollRowCapsContentHeight(t *testing.T) {
	row := ScrollRow()
	content := widget.NewGraphic(
		widget.GraphicOpts.Image(ebiten.NewImage(100, 800)),
		widget.GraphicOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.GridLayoutData{MaxHeight: 400})),
	)
	bar := widget.NewGraphic(widget.GraphicOpts.Image(ebiten.NewImage(10, 400)))
	row.AddChild(content, bar)
	_, height := row.PreferredSize()
	if height != 400 {
		t.Fatalf("row height = %d, want capped height 400", height)
	}
}
