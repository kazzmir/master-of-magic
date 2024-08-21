package util

import (
    "github.com/hajimehoshi/ebiten/v2"
)

type Animation struct {
    Frames []*ebiten.Image
    CurrentFrame int
    Loop bool
}

func MakeAnimation(frames []*ebiten.Image, loop bool) *Animation {
    return &Animation{
        Frames: frames,
        CurrentFrame: 0,
        Loop: loop,
    }
}

func (animation *Animation) Next() bool {
    if animation.CurrentFrame < len(animation.Frames) - 1 {
        animation.CurrentFrame += 1
        return true
    } else if animation.Loop {
        animation.CurrentFrame = 0
        return true
    } else {
        return false
    }
}

func (animation *Animation) Frame() *ebiten.Image {
    return animation.Frames[animation.CurrentFrame]
}
