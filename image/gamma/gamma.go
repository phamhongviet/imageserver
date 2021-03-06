package gamma

import (
	"image"
	"image/draw"
	"math"

	"github.com/pierrre/imageserver"
	imageserver_image "github.com/pierrre/imageserver/image"
	imageserver_image_internal "github.com/pierrre/imageserver/image/internal"
)

// Processor is a Processor that applies gamma transformation.
type Processor struct {
	vals        [1 << 16]uint16
	newDrawable func(image.Image) draw.Image
}

// NewProcessor creates a Processor.
func NewProcessor(gamma float64, highQuality bool) *Processor {
	prc := new(Processor)
	gammaInv := 1 / gamma
	for i := range prc.vals {
		prc.vals[i] = uint16(math.Pow(float64(i)/65535, gammaInv)*65535 + 0.5)
	}
	if highQuality {
		prc.newDrawable = func(p image.Image) draw.Image {
			return image.NewNRGBA64(image.Rect(0, 0, p.Bounds().Dx(), p.Bounds().Dy()))
		}
	} else {
		prc.newDrawable = imageserver_image_internal.NewDrawable
	}
	return prc
}

// Process implements Processor.
// It doesn't return error.
func (prc *Processor) Process(nim image.Image, params imageserver.Params) (image.Image, error) {
	width := nim.Bounds().Dx()
	height := nim.Bounds().Dy()
	dst := prc.newDrawable(nim)
	at := imageserver_image_internal.NewAtFunc(nim)
	set := imageserver_image_internal.NewSetFunc(dst)
	imageserver_image_internal.Parallel(height, func(yStart, yEnd int) {
		for y := yStart; y < yEnd; y++ {
			for x := 0; x < width; x++ {
				r, g, b, a := at(x, y)
				r, g, b, a = imageserver_image_internal.RGBAToNRGBA(r, g, b, a)
				r = uint32(prc.vals[uint16(r)])
				g = uint32(prc.vals[uint16(g)])
				b = uint32(prc.vals[uint16(b)])
				r, g, b, a = imageserver_image_internal.NRGBAToRGBA(r, g, b, a)
				set(x, y, r, g, b, a)
			}
		}
	})
	return dst, nil
}

// Change implements Processor.
func (prc *Processor) Change(params imageserver.Params) bool {
	return true
}

const correct = 2.2

// CorrectionProcessor is a Processor that corrects gamma.
type CorrectionProcessor struct {
	imageserver_image.Processor
	enabled bool
	before  *Processor
	after   *Processor
}

// NewCorrectionProcessor creates a CorrectionProcessor.
func NewCorrectionProcessor(prc imageserver_image.Processor, enabled bool) *CorrectionProcessor {
	return &CorrectionProcessor{
		Processor: prc,
		enabled:   enabled,
		before:    NewProcessor(1/correct, true),
		after:     NewProcessor(correct, true),
	}
}

// Process implements Processor.
func (prc *CorrectionProcessor) Process(nim image.Image, params imageserver.Params) (image.Image, error) {
	if !prc.Processor.Change(params) {
		return nim, nil
	}
	enabled, err := prc.isEnabled(params)
	if err != nil {
		return nil, err
	}
	if enabled {
		return prc.process(nim, params)
	}
	return prc.Processor.Process(nim, params)
}

func (prc *CorrectionProcessor) isEnabled(params imageserver.Params) (bool, error) {
	if params.Has("gamma_correction") {
		return params.GetBool("gamma_correction")
	}
	return prc.enabled, nil
}

func (prc *CorrectionProcessor) process(nim image.Image, params imageserver.Params) (image.Image, error) {
	original := nim
	nim, _ = prc.before.Process(nim, params)
	nim, err := prc.Processor.Process(nim, params)
	if err != nil {
		return nil, err
	}
	nim, _ = prc.after.Process(nim, params)
	if isHighQuality(nim) && !isHighQuality(original) {
		newNim := imageserver_image_internal.NewDrawableSize(original, nim.Bounds())
		imageserver_image_internal.Copy(newNim, nim)
		nim = newNim
	}
	return nim, nil
}

func isHighQuality(p image.Image) bool {
	switch p.(type) {
	case *image.RGBA64, *image.NRGBA64:
		return true
	default:
		return false
	}
}
