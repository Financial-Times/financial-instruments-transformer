package main

type fiService interface {
	Init(fit fiTransformer)
	Read(UUID string) (financialInstrument, bool)
	IDs() []string
	Count() int
	IsInitialised() bool
}

type fiServiceImpl struct {
	financialInstruments map[string]financialInstrument
}

func (fis *fiServiceImpl) Init(fit fiTransformer) {
	fis.financialInstruments = fit.Transform()
}

func (fis *fiServiceImpl) Read(UUID string) (financialInstrument, bool) {
	fi, present := fis.financialInstruments[UUID]
	return fi, present
}

func (fis *fiServiceImpl) IDs() []string {
	var UUIDs = []string{}
	for UUID := range fis.financialInstruments {
		UUIDs = append(UUIDs, UUID)
	}
	return UUIDs
}

func (fis *fiServiceImpl) Count() int {
	count := len(fis.financialInstruments)
	return count
}

func (fis *fiServiceImpl) IsInitialised() bool {
	if fis.financialInstruments == nil {
		return false
	}
	return true
}
