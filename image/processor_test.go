package image

import (
	"fmt"
	"image"
	"testing"

	"github.com/pierrre/imageserver"
)

var _ Processor = ProcessorFunc(nil)

func TestProcessorFunc(t *testing.T) {
	called := false
	f := ProcessorFunc(func(nim image.Image, params imageserver.Params) (image.Image, error) {
		called = true
		return nim, nil
	})
	nim := NewTestImage()
	f.Process(nim, imageserver.Params{})
	if !called {
		t.Fatal("not called")
	}
	if f.Change(imageserver.Params{}) != true {
		t.Fatal("not true")
	}
}

var _ Processor = ListProcessor{}

func TestListProcessorProcess(t *testing.T) {
	nim1 := NewTestImage()
	nim2 := NewTestImage()
	params := imageserver.Params{}
	prc := ListProcessor{}

	nim, err := prc.Process(nim1, params)
	if err != nil {
		t.Fatal(err)
	}
	if nim != nim1 {
		t.Fatal("not equal")
	}

	prc = append(prc, ProcessorFunc(func(image.Image, imageserver.Params) (image.Image, error) {
		return nim2, nil
	}))
	nim, err = prc.Process(nim1, params)
	if err != nil {
		t.Fatal(err)
	}
	if nim == nim1 {
		t.Fatal("equal")
	}
	if nim != nim2 {
		t.Fatal("not equal")
	}

	prc = append(prc, ProcessorFunc(func(image.Image, imageserver.Params) (image.Image, error) {
		return nil, fmt.Errorf("error")
	}))
	_, err = prc.Process(nim1, params)
	if err == nil {
		t.Fatal("no error")
	}
}

func TestListProcessorChange(t *testing.T) {
	params := imageserver.Params{}
	prc := ListProcessor{}

	if prc.Change(params) != false {
		t.Fatal("not false")
	}

	prc = append(prc, testChangeProcessor(false))
	if prc.Change(params) != false {
		t.Fatal("not false")
	}

	prc = append(prc, testChangeProcessor(true))
	if prc.Change(params) != true {
		t.Fatal("not true")
	}

	prc = append(prc, testChangeProcessor(false))
	if prc.Change(params) != true {
		t.Fatal("not true")
	}
}

type testChangeProcessor bool

func (prc testChangeProcessor) Process(nim image.Image, params imageserver.Params) (image.Image, error) {
	return nim, nil
}

func (prc testChangeProcessor) Change(params imageserver.Params) bool {
	return bool(prc)
}

var _ Processor = &ChangeProcessor{}

func TestChangeProcessor(t *testing.T) {
	prc := &ChangeProcessor{}
	change := prc.Change(imageserver.Params{})
	if change != true {
		t.Fatal("not false")
	}
}

var _ Processor = &IdentityProcessor{}

func TestIdentityProcessor(t *testing.T) {
	nim := NewTestImage()
	prc := &IdentityProcessor{}
	nimNew, err := prc.Process(nim, imageserver.Params{})
	if err != nil {
		t.Fatal(err)
	}
	if nimNew != nim {
		t.Fatal("not equal")
	}
	if prc.Change(imageserver.Params{}) != false {
		t.Fatal("not false")
	}
}
