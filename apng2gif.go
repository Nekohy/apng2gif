// Package apng2gif 把 APNG 转换为 GIF 的工具函数
package apng2gif

import (
	"fmt"
	"github.com/kettek/apng"
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"io"
	"math"
)

// Convert 把 r 中的 APNG 读出，转成 GIF 并写到 w
func Convert(r io.Reader, w io.Writer) error {
	anim, err := apng.DecodeAll(r)
	if err != nil {
		return err
	}
	if len(anim.Frames) < 2 {
		return fmt.Errorf("it is a normal PNG")
	}

	// 计算整幅画布大小
	full := computeCanvasRect(anim.Frames)

	// 帧合成
	imgs, delays, disposals, err := composeFrames(anim.Frames, full)
	if err != nil {
		return err
	}

	// 返回 GIF 数据
	return gif.EncodeAll(w, &gif.GIF{
		Image:    imgs,
		Delay:    delays,
		Disposal: disposals,
	})
}

// computeCanvasRect 返回所有帧覆盖后的主画板尺寸
func computeCanvasRect(frames []apng.Frame) image.Rectangle {
	var full image.Rectangle
	for _, f := range frames {
		r := frameRect(f)
		if full.Empty() {
			full = r
		} else {
			full = full.Union(r)
		}
	}
	return full
}

// composeFrames 合成帧并生成 GIF 需要的切片
func composeFrames(frames []apng.Frame, full image.Rectangle) ([]*image.Paletted, []int, []byte, error) {
	canvas := image.NewNRGBA(full)
	backup := image.NewNRGBA(full) // 用于 DisposePrevious

	var (
		gifImages    []*image.Paletted
		gifDelays    []int
		gifDisposals []byte
	)

	for _, f := range frames {
		target := frameRect(f)

		// BlendOp：BLEND_OP_SOURCE 擦除矩形后再绘
		if f.BlendOp == apng.BLEND_OP_SOURCE {
			clearRect(canvas, target)
		}

		// 绘制当前帧
		draw.Draw(canvas, target, f.Image, f.Image.Bounds().Min, draw.Over)

		// 整幅画布转成调色板图
		pal, err := convertToPaletted(canvas, full, nil)
		if err != nil {
			return nil, nil, nil, err
		}
		gifImages = append(gifImages, pal)
		gifDelays = append(gifDelays, int(math.Round(f.GetDelay()*100)))

		// DisposeOp
		switch f.DisposeOp {
		case apng.DISPOSE_OP_BACKGROUND:
			clearRect(canvas, target)
			gifDisposals = append(gifDisposals, gif.DisposalBackground)
		case apng.DISPOSE_OP_PREVIOUS:
			draw.Draw(canvas, full, backup, image.Point{}, draw.Src)
			gifDisposals = append(gifDisposals, gif.DisposalPrevious)
		default: // NONE
			gifDisposals = append(gifDisposals, gif.DisposalNone)
		}

		// 备份当前画布
		draw.Draw(backup, full, canvas, image.Point{}, draw.Src)
	}

	return gifImages, gifDelays, gifDisposals, nil
}

// frameRect 返回一帧在整幅画布中的绝对矩形
func frameRect(f apng.Frame) image.Rectangle {
	return image.Rect(
		f.XOffset,
		f.YOffset,
		f.XOffset+f.Image.Bounds().Dx(),
		f.YOffset+f.Image.Bounds().Dy(),
	)
}

// convertToPaletted 把任意 image 转成 *image.Paletted
func convertToPaletted(src image.Image, rect image.Rectangle, opt *gif.Options) (*image.Paletted, error) {
	var o gif.Options
	if opt != nil {
		o = *opt
	}
	if o.NumColors < 1 || o.NumColors > 256 {
		o.NumColors = 256
	}
	if o.Drawer == nil {
		o.Drawer = draw.FloydSteinberg
	}

	// 直接复用 paletted 图像，若颜色数符合要求
	if p, ok := src.(*image.Paletted); ok && len(p.Palette) <= o.NumColors {
		clone := image.NewPaletted(rect, p.Palette)
		o.Drawer.Draw(clone, rect, p, p.Rect.Min)
		clone.Palette[0] = color.RGBA{} // 让 0 号变透明
		return clone, nil
	}

	// 不符合则新建调色板图
	dst := image.NewPaletted(rect, palette.Plan9[:o.NumColors])
	dst.Palette[0] = color.RGBA{}
	o.Drawer.Draw(dst, rect, src, rect.Min)
	return dst, nil
}

// clearRect 把 NRGBA 区域擦为透明
func clearRect(img *image.NRGBA, r image.Rectangle) {
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			i := img.PixOffset(x, y)
			img.Pix[i+0] = 0
			img.Pix[i+1] = 0
			img.Pix[i+2] = 0
			img.Pix[i+3] = 0
		}
	}
}
