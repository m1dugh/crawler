package crawler

type Attachement struct{}

type Attachements map[string]*Attachement

func NewAttachements() Attachements {
	return make(Attachements)
}
