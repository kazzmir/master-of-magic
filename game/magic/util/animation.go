package util

import (
    "slices"
    "github.com/hajimehoshi/ebiten/v2"
)

type Animation struct {
    Frames []*ebiten.Image
    CurrentFrame int
    Repeat int
}

func MakeAnimation(frames []*ebiten.Image, loop bool) *Animation {
    if loop {
        return MakeRepeatAnimation(frames, -1)
    }
    return MakeRepeatAnimation(frames, 0)
}

func MakeReverseAnimation(frames []*ebiten.Image, loop bool) *Animation {
    newFrames := slices.Clone(frames)
    slices.Reverse(newFrames)

    return MakeAnimation(newFrames, loop)
}

func MakeRepeatAnimation(frames []*ebiten.Image, repeats int) *Animation {
    return &Animation{
        Frames: frames,
        CurrentFrame: 0,
        Repeat: repeats,
    }
}

/* returns true if there are animations left, otherwise false */
func (animation *Animation) Next() bool {
    if animation.CurrentFrame < len(animation.Frames) - 1 {
        animation.CurrentFrame += 1
        return true
    } else if animation.Repeat == -1 || animation.Repeat > 0 {
        animation.CurrentFrame = 0
        if animation.Repeat > 0 {
            animation.Repeat -= 1
        }
        return true
    } else {
        return false
    }
}

func (animation *Animation) Frame() *ebiten.Image {
    if animation.CurrentFrame < len(animation.Frames) {
        return animation.Frames[animation.CurrentFrame]
    }
    return nil
}
