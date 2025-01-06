package camera

const ZoomMin = 3
const ZoomMax = 10
const ZoomDefault = 10
const ZoomStep = 10

type Camera struct {
    // tile coordinates
    X int
    Y int
    // fractional tile coordinates
    DX float64
    DY float64

    Zoom int
    AnimatedZoom float64
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
    return ((float64(camera.Zoom) + camera.AnimatedZoom) / float64(ZoomStep))
}

func (camera *Camera) GetZoomedX() float64 {
    return camera.GetOffsetX() - 6.0 / camera.GetAnimatedZoom()
}

func (camera *Camera) GetZoomedY() float64 {
    return camera.GetOffsetY() - 5.0 / camera.GetAnimatedZoom()
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

func MakeCamera() Camera {
    return MakeCameraAt(0, 0)
}

func MakeCameraAt(x int, y int) Camera {
    return Camera{
        X: x,
        Y: y,
        Zoom: ZoomDefault,
        AnimatedZoom: 0,
    }
}
