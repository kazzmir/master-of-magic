package camera

const ZoomMin = 2
const ZoomMax = 12
const ZoomDefault = ZoomMax
// const ZoomStep = 15

type Camera struct {
    // tile coordinates
    X int
    Y int
    // fractional tile coordinates
    DX float64
    DY float64

    Zoom int
    AnimatedZoom float64

    SizeX int
    SizeY int
}

func (camera *Camera) SetOffset(x float64, y float64) {
    camera.DX = x
    camera.DY = y
}

func (camera *Camera) GetOffsetX() float64 {
    return float64(camera.X) + camera.DX
}

func (camera *Camera) GetOffsetY() float64 {
    return float64(camera.Y) + camera.DY
}

func (camera *Camera) GetZoom() float64 {
    return camera.GetAnimatedZoom()
    // return float64(camera.Zoom) / float64(ZoomStep)
}

func (camera *Camera) GetAnimatedZoom() float64 {
    return ((float64(camera.Zoom) + camera.AnimatedZoom) / float64(ZoomMax))
}

func (camera *Camera) GetZoomedX() float64 {
    return camera.GetOffsetX() - float64(camera.SizeX) / 2 / camera.GetAnimatedZoom()
}

func (camera *Camera) GetZoomedY() float64 {
    return camera.GetOffsetY() - float64(camera.SizeY) / 2 / camera.GetAnimatedZoom()
}

func (camera *Camera) GetZoomedMaxY() float64 {
    // FIXME: not sure why +1 is needed here. it doesn't fully solve the problem of there being
    // a gap at the bottom of the map sometimes
    return camera.GetOffsetY() + float64(camera.SizeY+1) / 2 / camera.GetAnimatedZoom()
}

func (camera *Camera) GetX() int {
    return camera.X
}

func (camera *Camera) GetY() int {
    return camera.Y
}

func (camera *Camera) Move(dx int, dy int) {
    camera.X += dx
    camera.Y += dy
}

func (camera *Camera) Center(x int, y int) {
    camera.X = x
    camera.Y = y

    if camera.Y < 0 {
        camera.Y = 0
    }
}

// return the bounds of a rectangle upper left (x1, y1) and lower right (x2, y2)
// all tiles within these bounds are visible, with some margin of error to account for edges
func (camera *Camera) GetTileBounds() (int, int, int, int) {
    minX := int(camera.GetZoomedX() - 1)
    minY := int(camera.GetZoomedY() - 1)
    // FIXME: 12 should be based on SizeX/SizeY
    maxX := minX + int(12/camera.GetZoom() + 3)
    maxY := minY + int(12/camera.GetZoom() + 3)

    return minX, minY, maxX, maxY
}

func (camera Camera) UpdateSize(sizeX int, sizeY int) Camera {
    camera.SizeX = sizeX
    camera.SizeY = sizeY
    return camera
}

func MakeCamera() Camera {
    return MakeCameraAt(0, 0)
}

func MakeCameraAt(x int, y int) Camera {
    return Camera{
        X: x,
        Y: y,
        // default size of screen
        SizeX: 12,
        SizeY: 10,
        Zoom: ZoomDefault,
        AnimatedZoom: 0,
    }
}
