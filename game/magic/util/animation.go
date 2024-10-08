package util

import (
    "slices"
    "github.com/hajimehoshi/ebiten/v2"
)

type Animation struct {
    Frames []*ebiten.Image
    CurrentFrame int
    Repeat int
    // done is true once Next() has been called after the last frame
    done bool
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

func (animation *Animation) Done() bool {
    return animation.done
}

/* returns true if there are frames left, otherwise false */
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
        animation.done = true
        return false
    }
}

func (animation *Animation) Frame() *ebiten.Image {
    if animation.CurrentFrame < len(animation.Frames) {
        return animation.Frames[animation.CurrentFrame]
    }
    return nil
}
