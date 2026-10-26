package engine

import "github.com/hajimehoshi/ebiten/v2"

func scaledSetupImage(source *ebiten.Image, width, height int) *ebiten.Image {
	image := ebiten.NewImage(width, height)
	scale := min(float64(width)/float64(source.Bounds().Dx()), float64(height)/float64(source.Bounds().Dy()))
	options := ebiten.DrawImageOptions{}
	options.GeoM.Scale(scale, scale)
	options.GeoM.Translate((float64(width)-float64(source.Bounds().Dx())*scale)/2, (float64(height)-float64(source.Bounds().Dy())*scale)/2)
	image.DrawImage(source, &options)
	return image
}

func (editor *gameSetupUI) loadBookIcons() {
	// spells.lbx symbols, ordered like Loadout.Books:
	// Life (ankh), Sorcery, Nature (tree), Death (skull), Chaos.
	indices := [5]int{7, 5, 4, 8, 6}
	loaded := 0
	for i, index := range indices {
		if source, err := editor.imageCache.GetImage("spells.lbx", index, 0); err == nil {
			editor.bookIcons[i].Image = scaledSetupImage(source, 24, 24)
			loaded++
		}
	}
	editor.bookIconsReady = loaded == len(indices)
}
