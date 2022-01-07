package crawler

// the keys of Attachements
type Attachements map[string]string

func NewAttachements() Attachements {
	return make(Attachements, 0)
}

type OnPageResultAdded func(
	body []byte,
	pageResults PageResult,
	domainResults DomainResultEntry,
) Attachements
