package crawler

type Attachement interface{}

type Attachements map[string]*Attachement

func NewAttachements() Attachements {
	return make(Attachements)
}
