package crawler

// the keys of Attachements
type Attachements map[string]string

func (att Attachements) AddAll(other Attachements) {
	for key, value := range other {
		att[key] = value
	}
}

func NewAttachements() Attachements {
	return make(Attachements, 0)
}

type OnPageResultAdded func(
	body []byte,
	pageResults PageResult,
	domainResults DomainResultEntry,
) Attachements
